package golang

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"ctxengine/internal/entity"

	sitter "github.com/smacker/go-tree-sitter"
	tsgolang "github.com/smacker/go-tree-sitter/golang"
)

// список встроенных функций Go для классификации вызовов
var goBuiltins = map[string]struct{}{
	"append": {}, "cap": {}, "close": {}, "complex": {}, "copy": {},
	"delete": {}, "imag": {}, "len": {}, "make": {}, "new": {},
	"panic": {}, "print": {}, "println": {}, "real": {}, "recover": {},
}

// ключевые слова Go
var goKeywords = []string{
	"break", "default", "func", "interface", "select",
	"case", "defer", "go", "map", "struct",
	"chan", "else", "goto", "package", "switch",
	"const", "fallthrough", "if", "range", "type",
	"continue", "for", "import", "return", "var",
}

// глобальная карта комментариев (по конечной строке блока комментариев)
var goComments map[int][]string

func isGoBuiltin(name string) bool {
	_, ok := goBuiltins[name]
	return ok
}

// ParseFile читает исходный код файла (source) и парсит его с помощью библиотеки go-tree-sitter.
// Результатом является структура FileCodeStruct, содержащая проиндексированный код, импорты, функции, структуры и т.д.
// Аргументы:
//
//	ctx: контекст для отмены операции парсинга;
//	fileName: путь к файлу, используется для сохранения пути в структуре;
//	source: исходное содержимое файла (в виде среза байт);
//	maxDepth: максимальная глубина рекурсивного обхода AST.
//	modulePath: модуль текущего проекта из go.mod для классификации импортов.
func ParseFile(ctx context.Context, fileName string, source []byte, maxDepth int, modulePath string) (*entity.FileCodeStruct, error) {
	// Создаём новый парсер и задаём язык Go.
	parser := sitter.NewParser()
	parser.SetLanguage(tsgolang.GetLanguage())

	// Парсим исходный код и получаем дерево синтаксического разбора (AST).
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
	rootNode := tree.RootNode()

	// Создаем структуру для хранения проиндексированного кода.
	fsStruct := &entity.FileCodeStruct{
		FileName:   fileName,
		Language:   entity.TypeFileGolang,
		Imports:    entity.ImportBlock{System: []string{}, External: []string{}, Internal: []string{}},
		Interfaces: []*entity.InterfaceInfo{},
		Variables:  []*entity.TokenInfo{},
		Structs:    []*entity.StructInfo{},
		Classes:    []*entity.ClassInfo{},
		Constants:  []*entity.ConstInfo{},
		Functions:  []*entity.FuncInfo{},
		Keywords:   []string{},
	}

	// Устанавливаем модуль и профиль языка
	fsStruct.ModulePath = modulePath
	fsStruct.Profile = &entity.LanguageProfile{
		Name:             entity.TypeFileGolang,
		TypeSystem:       entity.TypeSystemStatic,
		Paradigms:        []entity.ParadigmType{entity.ParadigmProcedural, entity.ParadigmFunctional, entity.ParadigmGeneric},
		Inheritance:      entity.InheritanceInterface,
		HasGenerics:      true,
		IsObjectOriented: true,
		IsFunctional:     true,
	}

	// Определяем имя пакета
	fsStruct.PackageName = extractPackageName(rootNode, source)

	// Собираем ключевые слова
	fsStruct.Keywords = collectKeywords(rootNode, source, goKeywords)

	// Собираем карту комментариев для последующей привязки
	goComments = collectComments(rootNode, source)

	// Начинаем рекурсивный обход дерева AST.
	// parentFunc == nil означает, что мы на верхнем уровне (не внутри функции).
	processNode(rootNode, source, fsStruct, nil, 0, maxDepth, fileName, modulePath)
	return fsStruct, nil
}

// collectKeywords собирает список ключевых слов, встреченных в тексте узлов AST (кроме строк и комментариев)
func collectKeywords(root *sitter.Node, source []byte, keywords []string) []string {
	if root == nil {
		return nil
	}
	kwset := map[string]struct{}{}
	skip := map[string]struct{}{
		"comment":                    {},
		"raw_string_literal":         {},
		"interpreted_string_literal": {},
		"string_literal":             {},
		"rune_literal":               {},
	}
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n == nil {
			return
		}
		if _, ok := skip[n.Type()]; ok {
			return
		}
		text := string(n.Content(source))
		for _, kw := range keywords {
			if strings.Contains(text, kw) {
				kwset[kw] = struct{}{}
			}
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	var out []string
	for k := range kwset {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// collectComments извлекает комментарии и группирует по конечной строке
func collectComments(root *sitter.Node, source []byte) map[int][]string {
	result := map[int][]string{}
	if root == nil {
		return result
	}
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n.Type() == "comment" {
			endLine := int(n.EndPoint().Row) + 1
			text := string(n.Content(source))
			lines := strings.Split(text, "\n")
			for i := range lines {
				lines[i] = strings.TrimSpace(lines[i])
			}
			result[endLine] = append(result[endLine], lines...)
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(root)
	return result
}

// getLeadingComments возвращает комментарии, непосредственно предшествующие строке startLine
func getLeadingComments(startLine int, comments map[int][]string) []string {
	var res []string
	for l := startLine - 1; l >= 1; l-- {
		if c, ok := comments[l]; ok {
			res = append(c, res...)
			continue
		}
		break
	}
	return res
}

// extractPackageName извлекает имя пакета из корня AST (узел package_clause)
func extractPackageName(root *sitter.Node, source []byte) string {
	if root == nil {
		return ""
	}
	for i := 0; i < int(root.ChildCount()); i++ {
		ch := root.Child(i)
		if ch.Type() == "package_clause" {
			for j := 0; j < int(ch.ChildCount()); j++ {
				id := ch.Child(j)
				if id.Type() == "identifier" {
					return string(id.Content(source))
				}
			}
		}
	}
	return ""
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
//	modulePath: модуль текущего проекта из go.mod для классификации импортов.
func processNode(node *sitter.Node, source []byte, fs *entity.FileCodeStruct, parentFunc *entity.FuncInfo, depth, maxDepth int, fileName string, modulePath string) {
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
		// Привязываем ведущие комментарии к объявлению функции
		funcInfo.TokenInfo.Comments = getLeadingComments(int(node.StartPoint().Row)+1, goComments)
		// Попробуем определить владельца/получатель для метода (receiver)
		if nodeType == "method_declaration" {
			// В go-tree-sitter у метода есть поле receiver (field: receiver)
			rcv := node.ChildByFieldName("receiver")
			if rcv != nil {
				recvText := string(rcv.Content(source))
				funcInfo.ReceiverType = strings.TrimSpace(recvText)
				funcInfo.OwnerName = funcInfo.ReceiverType
				funcInfo.OwnerKind = "struct"
			}
		}
		// Вычисляем FQN (пакет[.тип]::имя)
		pkg := fs.PackageName
		owner := funcInfo.OwnerName
		if owner != "" {
			funcInfo.FQN = fmt.Sprintf("%s.%s::%s", pkg, owner, funcInfo.Name)
		} else {
			funcInfo.FQN = fmt.Sprintf("%s::%s", pkg, funcInfo.Name)
		}
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
			callInfo := processCallExpression(node, source, fs)
			if callInfo != nil {
				parentFunc.Functions = append(parentFunc.Functions, callInfo)
			}
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
			parseImports(node, source, fs, modulePath)
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
		processNode(child, source, fs, currentParent, depth+1, maxDepth, fileName, modulePath)
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

	// Ищем имя функции: identifier или field_identifier
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "field_identifier" {
			funcInfo.Name = string(child.Content(source))
			break
		}
	}

	// Ищем узел signature и извлекаем параметры/результаты из него
	var signature *sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "signature" {
			signature = child
			break
		}
	}
	if signature != nil {
		for i := 0; i < int(signature.ChildCount()); i++ {
			ch := signature.Child(i)
			switch ch.Type() {
			case "parameter_list":
				funcInfo.Parameters = processParamList(ch, source)
			case "result":
				if ch.ChildCount() == 1 && ch.Child(0).Type() != "parameter_list" {
					funcInfo.Returns = []*entity.ParamInfo{{Name: "", Type: string(ch.Content(source))}}
				} else {
					funcInfo.Returns = processParamList(ch, source)
				}
			}
		}
	} else {
		// На случай, если parser выдаёт parameter_list/result прямо под function_declaration
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			switch child.Type() {
			case "parameter_list":
				funcInfo.Parameters = processParamList(child, source)
			case "result":
				if child.ChildCount() == 1 && child.Child(0).Type() != "parameter_list" {
					funcInfo.Returns = []*entity.ParamInfo{{Name: "", Type: string(child.Content(source))}}
				} else {
					funcInfo.Returns = processParamList(child, source)
				}
			}
		}
	}

	// Сохраняем полный текст объявления функции.
	funcInfo.Content = string(source[node.StartByte():node.EndByte()])
	// Комментарии перед функцией
	funcInfo.TokenInfo.Comments = getLeadingComments(int(node.StartPoint().Row)+1, goComments)
	return funcInfo
}

// processCallExpression обрабатывает узел вызова функции и возвращает объект FuncInfo,
// содержащий информацию о вызове, его аргументах и исходном коде вызова.
func processCallExpression(node *sitter.Node, source []byte, fs *entity.FileCodeStruct) *entity.FuncInfo {
	callInfo := &entity.FuncInfo{
		TokenInfo: entity.TokenInfo{
			StartLine: int(node.StartPoint().Row) + 1,
			EndLine:   int(node.EndPoint().Row) + 1,
		},
		Parameters: []*entity.ParamInfo{},
		Returns:    []*entity.ParamInfo{},
		Functions:  []*entity.FuncInfo{},
		IsCall:     true,
	}

	// Определяем имя вызова и его вид: identifier (function/builtin) или selector_expression (method/pkg.Func)
	isSelector := false
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "selector_expression" {
			callInfo.Name = string(child.Content(source))
			isSelector = true
			break
		}
		if child.Type() == "identifier" {
			callInfo.Name = string(child.Content(source))
		}
	}

	// Если имя не удалось определить (например, это приведение типа), пропускаем добавление как вызова функции
	if len(callInfo.Name) == 0 {
		return nil
	}

	// Классификация: selector_expression вида pkg.Func → function, obj.Method → method.
	if isSelector {
		// Попробуем распарсить selector_expression: left.right
		parts := strings.Split(callInfo.Name, ".")
		if len(parts) >= 2 {
			left := parts[0]
			// Если левая часть равна имени пакета текущего файла или совпадает с импортированным пакетом — это функция пакета
			isPkg := left == fs.PackageName
			if !isPkg {
				// грубая эвристика: если left встречается в списке imports (последняя часть пути), считаем пакетом
				for _, imp := range append(append([]string{}, fs.Imports.System...), append(fs.Imports.External, fs.Imports.Internal...)...) {
					last := imp
					if idx := strings.LastIndex(imp, "/"); idx >= 0 {
						last = imp[idx+1:]
					}
					if last == left {
						isPkg = true
						break
					}
				}
			}
			if isPkg {
				callInfo.CallKind = entity.CallKindFunction
			} else {
				callInfo.CallKind = entity.CallKindMethod
			}
		} else {
			callInfo.CallKind = entity.CallKindMethod
		}
	} else {
		if isGoBuiltin(callInfo.Name) {
			callInfo.CallKind = entity.CallKindBuiltin
		} else {
			callInfo.CallKind = entity.CallKindFunction
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
			Name:  "",                            // имя аргумента в вызове неизвестно
			Type:  "",                            // тип в выражении вызова не определяем здесь
			Value: string(child.Content(source)), // исходное выражение аргумента
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
		// Вложенный список параметров
		if child.Type() == "parameter_list" {
			nested := processParamList(child, source)
			params = append(params, nested...)
			continue
		}
		if child.Type() == "parameter_declaration" {
			param := &entity.ParamInfo{}
			var typeParts []string
			// В каждом параметре ищем узлы, содержащие имя и тип.
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "," || subChild.Type() == "(" || subChild.Type() == ")" {
					continue
				}
				if subChild.Type() == "identifier" && param.Name == "" {
					param.Name = string(subChild.Content(source))
					continue
				}
				typeParts = append(typeParts, string(subChild.Content(source)))
			}
			if len(typeParts) > 0 {
				param.Type = strings.TrimSpace(strings.Join(typeParts, " "))
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
	// Привяжем ведущие комментарии
	token.Comments = getLeadingComments(token.StartLine, goComments)
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
			constInfo.Comments = getLeadingComments(int(child.StartPoint().Row)+1, goComments)
			consts = append(consts, constInfo)
		}
	}
	return consts
}

// parseImports рекурсивно ищет узлы "import_spec" в пределах узла "import_declaration".
// Для каждого найденного импорта извлекается путь, и он добавляется в соответствующий список:
// системные импорты (если путь не содержит точки) или внешние импорты (если содержит точку).
// Теперь также поддерживается Internal (импорты текущего модуля).
func parseImports(node *sitter.Node, source []byte, fs *entity.FileCodeStruct, modulePath string) {
	var search func(n *sitter.Node)
	search = func(n *sitter.Node) {
		if n.Type() == "import_spec" {
			// Извлекаем узел с именем поля "path".
			pathNode := n.ChildByFieldName("path")
			if pathNode != nil {
				// Обрезаем кавычки вокруг пути.
				pathText := strings.Trim(string(pathNode.Content(source)), `"`)
				// Классификация: Internal > External > System
				if modulePath != "" && (pathText == modulePath || strings.HasPrefix(pathText, modulePath+"/")) {
					fs.Imports.Internal = append(fs.Imports.Internal, pathText)
				} else if strings.Contains(pathText, ".") {
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
			// Имя типа
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				if subChild.Type() == "type_identifier" || subChild.Type() == "identifier" {
					typeName = string(subChild.Content(source))
					break
				}
			}
			// Поиск определения структуры или интерфейса
			for j := 0; j < int(child.ChildCount()); j++ {
				subChild := child.Child(j)
				switch subChild.Type() {
				case "struct_type":
					structInfo := &entity.StructInfo{
						TokenInfo: entity.TokenInfo{
							StartLine: int(subChild.StartPoint().Row) + 1,
							EndLine:   int(subChild.EndPoint().Row) + 1,
						},
						Name:     typeName,
						Comments: []string{},
						Fields:   make(map[string]*entity.FieldInfo),
					}
					structInfo.FQN = fmt.Sprintf("%s.%s", fs.PackageName, structInfo.Name)
					// комментарии к объявлению типа (берём по началу type_spec)
					structInfo.TokenInfo.Comments = getLeadingComments(int(child.StartPoint().Row)+1, goComments)
					// Грубый разбор полей структуры
					for k := 0; k < int(subChild.ChildCount()); k++ {
						fld := subChild.Child(k)
						if fld.Type() == "field_declaration" {
							var fieldNames []string
							var typeText string
							for m := 0; m < int(fld.ChildCount()); m++ {
								p := fld.Child(m)
								if p.Type() == "identifier" {
									fieldNames = append(fieldNames, string(p.Content(source)))
									continue
								}
								if p.Type() == ":" || p.Type() == "," || p.Type() == "tag" {
									continue
								}
								if len(typeText) == 0 {
									typeText = string(p.Content(source))
								}
							}
							for _, fname := range fieldNames {
								structInfo.Fields[fname] = &entity.FieldInfo{
									TokenInfo: entity.TokenInfo{
										StartLine: int(fld.StartPoint().Row) + 1,
										EndLine:   int(fld.EndPoint().Row) + 1,
									},
									Name: fname,
									Type: typeText,
								}
							}
						}
					}
					fs.Structs = append(fs.Structs, structInfo)
				case "interface_type":
					iface := &entity.InterfaceInfo{
						TokenInfo: entity.TokenInfo{
							StartLine: int(subChild.StartPoint().Row) + 1,
							EndLine:   int(subChild.EndPoint().Row) + 1,
						},
						Name:    typeName,
						Methods: make(map[string]*entity.FuncInfo),
					}
					iface.FQN = fmt.Sprintf("%s.%s", fs.PackageName, iface.Name)
					iface.TokenInfo.Comments = getLeadingComments(int(child.StartPoint().Row)+1, goComments)
					// Разбор method_spec
					for k := 0; k < int(subChild.ChildCount()); k++ {
						ms := subChild.Child(k)
						if ms.Type() == "method_spec" {
							method := &entity.FuncInfo{
								TokenInfo: entity.TokenInfo{
									StartLine: int(ms.StartPoint().Row) + 1,
									EndLine:   int(ms.EndPoint().Row) + 1,
								},
								Parameters: []*entity.ParamInfo{},
								Returns:    []*entity.ParamInfo{},
							}
							// имя метода
							for m := 0; m < int(ms.ChildCount()); m++ {
								c := ms.Child(m)
								if c.Type() == "identifier" || c.Type() == "field_identifier" {
									method.Name = string(c.Content(source))
									break
								}
							}
							// параметры и возвращаемые значения
							for m := 0; m < int(ms.ChildCount()); m++ {
								c := ms.Child(m)
								if c.Type() == "parameter_list" {
									method.Parameters = processParamList(c, source)
								} else if c.Type() == "result" {
									if c.ChildCount() == 1 && c.Child(0).Type() != "parameter_list" {
										method.Returns = []*entity.ParamInfo{{Name: "", Type: string(c.Content(source))}}
									} else {
										method.Returns = processParamList(c, source)
									}
								}
							}
							method.Content = string(source[ms.StartByte():ms.EndByte()])
							iface.Methods[method.Name] = method
						}
					}
					fs.Interfaces = append(fs.Interfaces, iface)
				}
			}
		}
	}
}
