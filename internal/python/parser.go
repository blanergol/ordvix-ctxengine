package python

import (
	"context"
	"sort"
	"strings"

	"ctxengine/internal/entity"

	sitter "github.com/smacker/go-tree-sitter"
	tspython "github.com/smacker/go-tree-sitter/python"
)

// ParseFile парсит Python-файл и возвращает унифицированную структуру индекса.
func ParseFile(ctx context.Context, fileName string, source []byte, maxDepth int, modulePath string) (*entity.FileCodeStruct, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(tspython.GetLanguage())

	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
	root := tree.RootNode()

	fs := &entity.FileCodeStruct{
		FileName:   fileName,
		Language:   entity.TypeFilePython,
		Imports:    entity.ImportBlock{System: []string{}, External: []string{}, Internal: []string{}},
		Interfaces: []*entity.InterfaceInfo{},
		Variables:  []*entity.TokenInfo{},
		Structs:    []*entity.StructInfo{},
		Classes:    []*entity.ClassInfo{},
		Constants:  []*entity.ConstInfo{},
		Functions:  []*entity.FuncInfo{},
		Keywords:   []string{},
	}

	// Профиль языка
	fs.Profile = &entity.LanguageProfile{
		Name:             entity.TypeFilePython,
		TypeSystem:       entity.TypeSystemDynamic,
		Paradigms:        []entity.ParadigmType{entity.ParadigmOOP, entity.ParadigmFunctional, entity.ParadigmProcedural},
		Inheritance:      entity.InheritanceClass,
		HasGenerics:      false,
		IsObjectOriented: true,
		IsFunctional:     true,
	}

	// Ключевые слова Python
	pyKeywords := []string{"False", "None", "True", "and", "as", "assert", "async", "await", "break", "class", "continue", "def", "del", "elif", "else", "except", "finally", "for", "from", "global", "if", "import", "in", "is", "lambda", "nonlocal", "not", "or", "pass", "raise", "return", "try", "while", "with", "yield"}
	fs.Keywords = collectKeywords(root, source, pyKeywords)

	processNode(root, source, fs, nil, 0, maxDepth)
	return fs, nil
}

func processNode(node *sitter.Node, source []byte, fs *entity.FileCodeStruct, parentFunc *entity.FuncInfo, depth, maxDepth int) {
	if depth > maxDepth {
		return
	}
	t := node.Type()
	currentParent := parentFunc

	switch t {
	case "function_definition":
		f := processFunction(node, source, fs.FileName)
		if parentFunc != nil {
			parentFunc.Functions = append(parentFunc.Functions, f)
		} else {
			fs.Functions = append(fs.Functions, f)
		}
		currentParent = f
	case "call":
		if parentFunc != nil {
			c := processCall(node, source)
			if c != nil {
				parentFunc.Functions = append(parentFunc.Functions, c)
			}
		}
	case "class_definition":
		cls := processClass(node, source)
		fs.Classes = append(fs.Classes, cls)
	case "import_statement", "import_from_statement":
		parseImports(node, source, fs)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		processNode(node.Child(i), source, fs, currentParent, depth+1, maxDepth)
	}
}

func processFunction(node *sitter.Node, source []byte, fileName string) *entity.FuncInfo {
	fi := &entity.FuncInfo{
		TokenInfo:  entity.TokenInfo{StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1},
		FileName:   fileName,
		Parameters: []*entity.ParamInfo{},
		Returns:    []*entity.ParamInfo{},
		Functions:  []*entity.FuncInfo{},
	}
	// имя после ключевого слова def
	for i := 0; i < int(node.ChildCount()); i++ {
		ch := node.Child(i)
		if ch.Type() == "identifier" {
			fi.Name = string(ch.Content(source))
			break
		}
	}
	fi.Content = string(source[node.StartByte():node.EndByte()])
	return fi
}

func processCall(node *sitter.Node, source []byte) *entity.FuncInfo {
	ci := &entity.FuncInfo{
		TokenInfo:  entity.TokenInfo{StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1},
		Parameters: []*entity.ParamInfo{},
		Returns:    []*entity.ParamInfo{},
		Functions:  []*entity.FuncInfo{},
		IsCall:     true,
	}
	// имя вызова в python AST: attribute (obj.method) или identifier
	for i := 0; i < int(node.ChildCount()); i++ {
		ch := node.Child(i)
		if ch.Type() == "attribute" || ch.Type() == "identifier" {
			ci.Name = string(ch.Content(source))
			if ch.Type() == "attribute" {
				ci.CallKind = entity.CallKindMethod
			} else {
				// упрощённо: без списка builtins — считаем пользовательской функцией
				ci.CallKind = entity.CallKindFunction
			}
			break
		}
	}
	if len(ci.Name) == 0 {
		return nil
	}
	ci.Content = string(source[node.StartByte():node.EndByte()])
	return ci
}

func processClass(node *sitter.Node, source []byte) *entity.ClassInfo {
	cls := &entity.ClassInfo{
		TokenInfo: entity.TokenInfo{StartLine: int(node.StartPoint().Row) + 1, EndLine: int(node.EndPoint().Row) + 1},
		Name:      "",
		Fields:    map[string]*entity.FieldInfo{},
		Methods:   map[string]*entity.FuncInfo{},
	}
	// имя класса после ключевого слова class
	for i := 0; i < int(node.ChildCount()); i++ {
		ch := node.Child(i)
		if ch.Type() == "identifier" {
			cls.Name = string(ch.Content(source))
			break
		}
	}
	// грубый разбор полей/методов внутри suite
	for i := 0; i < int(node.ChildCount()); i++ {
		ch := node.Child(i)
		if ch.Type() == "block" || ch.Type() == "suite" {
			for j := 0; j < int(ch.ChildCount()); j++ {
				m := ch.Child(j)
				if m.Type() == "function_definition" {
					fn := processFunction(m, source, "")
					cls.Methods[fn.Name] = fn
				} else if m.Type() == "expression_statement" {
					// возможные атрибуты вида self.x = ...
					text := string(m.Content(source))
					if strings.Contains(text, "self.") && strings.Contains(text, "=") {
						// очень грубо извлечём имя после self.
						idx := strings.Index(text, "self.")
						rest := text[idx+5:]
						name := rest
						if p := strings.IndexAny(rest, " =\t\n"); p >= 0 {
							name = rest[:p]
						}
						cls.Fields[name] = &entity.FieldInfo{Name: name}
					}
				}
			}
		}
	}
	return cls
}

func parseImports(node *sitter.Node, source []byte, fs *entity.FileCodeStruct) {
	var walk func(n *sitter.Node)
	walk = func(n *sitter.Node) {
		if n.Type() == "dotted_name" || n.Type() == "aliased_import" || n.Type() == "import_from_statement" {
			path := strings.TrimSpace(string(n.Content(source)))
			if strings.Contains(path, ".") {
				fs.Imports.External = append(fs.Imports.External, path)
			} else {
				fs.Imports.System = append(fs.Imports.System, path)
			}
		}
		for i := 0; i < int(n.ChildCount()); i++ {
			walk(n.Child(i))
		}
	}
	walk(node)
}

// collectKeywords обходит AST и извлекает использованные ключевые слова, игнорируя строки/комментарии
func collectKeywords(root *sitter.Node, source []byte, keywords []string) []string {
	if root == nil {
		return nil
	}
	kwset := map[string]struct{}{}
	skip := map[string]struct{}{
		"comment":        {},
		"string":         {},
		"string_content": {},
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
