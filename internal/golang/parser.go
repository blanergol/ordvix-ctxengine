package golang

import (
	"context"
	"strings"

	"ctxengine/internal/entity"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
)

// ParseFile читает исходный код файла (source) и парсит его с помощью библиотеки go-tree-sitter.
// Результатом является структура FileCodeStruct, содержащая проиндексированный код, импорты, функции, структуры и т.д.
// Аргументы:
//
//	ctx: контекст для отмены операции парсинга;
//	fileName: путь к файлу, используется для сохранения пути в структуре;
//	source: исходное содержимое файла (в виде среза байт);
//	maxDepth: максимальная глубина рекурсивного обхода AST.
func ParseFile(ctx context.Context, fileName string, source []byte, maxDepth int) (*entity.FileCodeStruct, error) {
	// Создаём новый парсер и задаём язык Go.
	parser := sitter.NewParser()
	parser.SetLanguage(golang.GetLanguage())

	// Парсим исходный код и получаем дерево синтаксического разбора (AST).
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
	rootNode := tree.RootNode()

	// Создаем структуру для хранения проиндексированного кода.
	fsStruct := &entity.FileCodeStruct{
		FileName:   fileName,
		Imports:    entity.ImportBlock{System: []string{}, External: []string{}},
		Interfaces: []*entity.InterfaceInfo{},
		Variables:  []*entity.TokenInfo{},
		Structs:    []*entity.StructInfo{},
		Constants:  []*entity.ConstInfo{},
		Functions:  []*entity.FuncInfo{},
	}

	// Начинаем рекурсивный обход дерева AST.
	// parentFunc == nil означает, что мы на верхнем уровне (не внутри функции).
	processNode(rootNode, source, fsStruct, nil, 0, maxDepth, fileName)
	return fsStruct, nil
}

// processNode рекурсивно обходит дерево AST и делегирует обработку узлов в зависимости от их типа.
// Аргументы:
//
//	node: текущий узел AST;
//	source: исходный код файла;
//	fs: структура FileCodeStruct, куда добавляются результаты парсинга;
//	parentFunc: если не nil, указывает, что мы находимся внутри функции – вызовы будут добавляться как дочерние функции;
//	depth: текущая глубина рекурсии;
//	maxDepth: максимальная глубина обхода;
//	fileName: путь к файлу, используется для сохранения в найденных функциях.
func processNode(node *sitter.Node, source []byte, fs *entity.FileCodeStruct, parentFunc *entity.FuncInfo, depth, maxDepth int, fileName string) {
	if depth > maxDepth {
		return
	}

	// Получаем тип текущего узла.
	nodeType := node.Type()
	// currentParent будет передаваться в рекурсивном вызове – если мы внутри функции, то это родительская функция.
	currentParent := parentFunc

	switch nodeType {
	// Если узел является объявлением функции или метода:
	case "function_declaration", "method_declaration":
		// Обрабатываем объявление функции, создавая объект FuncInfo.
		funcInfo := processFunction(node, source, fileName)
		// Если мы уже находимся внутри функции, добавляем новую функцию как вложенную,
		// иначе – как верхнеуровневую (в FileCodeStruct).
		if parentFunc != nil {
			parentFunc.Functions = append(parentFunc.Functions, funcInfo)
		} else {
			fs.Functions = append(fs.Functions, funcInfo)
		}
		// Обновляем текущего родителя для последующего обхода – теперь это только что найденная функция.
		currentParent = funcInfo

	// Если узел представляет вызов функции:
	case "call_expression":
		// Если мы внутри функции, обрабатываем вызов и добавляем его как дочернюю функцию.
		if parentFunc != nil {
			callInfo := processCallExpression(node, source)
			parentFunc.Functions = append(parentFunc.Functions, callInfo)
		}

	// Если узел представляет объявление переменных вне функций:
	case "var_declaration":
		if parentFunc == nil {
			vars := processVarDeclaration(node)
			fs.Variables = append(fs.Variables, vars...)
		}

	// Если узел содержит объявление констант:
	case "const_declaration":
		if parentFunc == nil {
			consts := processConstDeclaration(node, source)
			fs.Constants = append(fs.Constants, consts...)
		}

	// Если узел представляет импорт:
	case "import_declaration":
		if parentFunc == nil {
			parseImports(node, source, fs)
		}

	// Если узел содержит объявление типа (структуры или интерфейса):
	case "type_declaration":
		if parentFunc == nil {
			processTypeDeclaration(node, source, fs)
		}
	}

	// Рекурсивно обрабатываем все дочерние узлы текущего узла.
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		processNode(child, source, fs, currentParent, depth+1, maxDepth, fileName)
	}
}

// processFunction обрабатывает узел объявления функции или метода, создавая объект FuncInfo.
// Сохраняет также путь к файлу, где определена функция.
func processFunction(node *sitter.Node, source []byte, fileName string) *entity.FuncInfo {
	// Создаем FuncInfo с информацией о начальной и конечной строках.
	funcInfo := &entity.FuncInfo{
		TokenInfo: entity.TokenInfo{
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
		},
		FileName:   fileName, // Сохраняем путь к файлу
		Parameters: []*entity.ParamInfo{},
		Returns:    []*entity.ParamInfo{},
		Functions:  []*entity.FuncInfo{},
	}

	// Ищем дочерний узел с типом "identifier" для получения имени функции.
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			funcInfo.Name = string(child.Content(source))
			break
		}
	}

	// Обрабатываем узлы, содержащие список параметров ("parameter_list") и возвращаемых значений ("result").
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "parameter_list":
			funcInfo.Parameters = processParamList(child, source)
		case "result":
			funcInfo.Returns = processParamList(child, source)
		}
	}

	// Сохраняем полный текст объявления функции.
	funcInfo.Content = string(source[node.StartByte():node.EndByte()])
	return funcInfo
}

// processCallExpression обрабатывает узел вызова функции и возвращает объект FuncInfo,
// содержащий информацию о вызове, его аргументах и исходном коде вызова.
func processCallExpression(node *sitter.Node, source []byte) *entity.FuncInfo {
	callInfo := &entity.FuncInfo{
		TokenInfo: entity.TokenInfo{
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
		},
		Parameters: []*entity.ParamInfo{},
		Returns:    []*entity.ParamInfo{},
		Functions:  []*entity.FuncInfo{},
	}

	// Определяем имя вызова: ищем узлы с типом "identifier" или "selector_expression".
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "selector_expression" {
			callInfo.Name = string(child.Content(source))
			break
		}
	}

	// Обрабатываем аргументы вызова, если узел имеет тип "argument_list".
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "argument_list" {
			callInfo.Parameters = processCallArguments(child, source)
		}
	}

	// Сохраняем текст вызова функции.
	callInfo.Content = string(source[node.StartByte():node.EndByte()])
	return callInfo
}

// processCallArguments обрабатывает узел "argument_list" вызова функции
// и возвращает список аргументов в виде среза ParamInfo.
func processCallArguments(node *sitter.Node, source []byte) []*entity.ParamInfo {
	var args []*entity.ParamInfo
	// Проходим по всем дочерним узлам и создаем параметр для каждого, который не является разделителем.
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		// Игнорируем разделители: запятые и скобки.
		if child.Type() == "," || child.Type() == "(" || child.Type() == ")" {
			continue
		}
		arg := &entity.ParamInfo{
			Name: "",                            // Имя аргумента не всегда можно определить, поэтому оставляем пустым.
			Type: string(child.Content(source)), // Сохраняем текстовое представление аргумента.
		}
		args = append(args, arg)
	}
	return args
}

// processParamList извлекает список параметров или возвращаемых значений из узла,
// который представляет список параметров (тип "parameter_list") или результаты ("result").
// Возвращается срез объектов ParamInfo.
func processParamList(node *sitter.Node, source []byte) []*entity.ParamInfo {
	var params []*entity.ParamInfo
	// Проходим по всем дочерним узлам, которые представляют отдельные параметры.
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "parameter_declaration" {
			param := &entity.ParamInfo{}
			// В каждом параметре ищем узлы, содержащие имя и тип.
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "identifier" && param.Name == "" {
					param.Name = string(subChild.Content(source))
				} else if param.Type == "" {
					param.Type = string(subChild.Content(source))
				}
			}
			params = append(params, param)
		}
	}
	return params
}

// processVarDeclaration извлекает переменные, объявленные вне функций, и возвращает их в виде среза TokenInfo.
func processVarDeclaration(node *sitter.Node) []*entity.TokenInfo {
	var tokens []*entity.TokenInfo
	token := &entity.TokenInfo{
		StartLine: int(node.StartPoint().Row) + 1,
		EndLine:   int(node.EndPoint().Row) + 1,
	}
	tokens = append(tokens, token)
	return tokens
}

// processConstDeclaration извлекает глобальные константы из узла "const_declaration" и возвращает срез объектов ConstInfo.
func processConstDeclaration(node *sitter.Node, source []byte) []*entity.ConstInfo {
	var consts []*entity.ConstInfo
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "const_spec" {
			constInfo := &entity.ConstInfo{
				TokenInfo: entity.TokenInfo{
					StartLine: int(child.StartPoint().Row) + 1,
					EndLine:   int(child.EndPoint().Row) + 1,
				},
				Value: string(child.Content(source)),
			}
			consts = append(consts, constInfo)
		}
	}
	return consts
}

// parseImports рекурсивно ищет узлы "import_spec" в пределах узла "import_declaration".
// Для каждого найденного импорта извлекается путь, и он добавляется в соответствующий список:
// системные импорты (если путь не содержит точки) или внешние импорты (если содержит точку).
func parseImports(node *sitter.Node, source []byte, fs *entity.FileCodeStruct) {
	var search func(n *sitter.Node)
	search = func(n *sitter.Node) {
		if n.Type() == "import_spec" {
			// Извлекаем узел с именем поля "path".
			pathNode := n.ChildByFieldName("path")
			if pathNode != nil {
				// Обрезаем кавычки вокруг пути.
				pathText := strings.Trim(string(pathNode.Content(source)), `"`)
				if strings.Contains(pathText, ".") {
					fs.Imports.External = append(fs.Imports.External, pathText)
				} else {
					fs.Imports.System = append(fs.Imports.System, pathText)
				}
			}
		}
		// Рекурсивно обходим все дочерние узлы.
		for i := 0; i < int(n.ChildCount()); i++ {
			search(n.Child(i))
		}
	}
	search(node)
}

// processTypeDeclaration обрабатывает объявления типов (структуры и интерфейсы).
// Для каждого узла "type_spec" извлекается имя типа, а затем в зависимости от типа определяется,
// является ли он структурой или интерфейсом, и соответствующая информация сохраняется в FileCodeStruct.
func processTypeDeclaration(node *sitter.Node, source []byte, fs *entity.FileCodeStruct) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "type_spec" {
			var typeName string
			// Ищем узел, содержащий имя типа.
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "type_identifier" || subChild.Type() == "identifier" {
					typeName = string(subChild.Content(source))
					break
				}
			}
			// Обрабатываем узлы, содержащие определение структуры или интерфейса.
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				switch subChild.Type() {
				case "struct_type":
					// Создаем объект структуры.
					structInfo := &entity.StructInfo{
						TokenInfo: entity.TokenInfo{
							StartLine: int(subChild.StartPoint().Row) + 1,
							EndLine:   int(subChild.EndPoint().Row) + 1,
						},
						Comments: []string{},
						Fields:   make(map[string]*entity.FieldInfo),
					}
					// Сохраняем имя поля как имя типа.
					structInfo.Fields[typeName] = &entity.FieldInfo{
						TokenInfo: entity.TokenInfo{
							StartLine: int(subChild.StartPoint().Row) + 1,
							EndLine:   int(subChild.EndPoint().Row) + 1,
						},
						Name: typeName,
						Type: string(subChild.Content(source)),
					}
					fs.Structs = append(fs.Structs, structInfo)
				case "interface_type":
					// Создаем объект интерфейса.
					iface := &entity.InterfaceInfo{
						TokenInfo: entity.TokenInfo{
							StartLine: int(subChild.StartPoint().Row) + 1,
							EndLine:   int(subChild.EndPoint().Row) + 1,
						},
						Methods: make(map[string]*entity.FuncInfo),
					}
					// Сохраняем метод с именем типа и текстом определения интерфейса.
					iface.Methods[typeName] = &entity.FuncInfo{
						Name:    typeName,
						Content: string(subChild.Content(source)),
					}
					fs.Interfaces = append(fs.Interfaces, iface)
				}
			}
		}
	}
}
