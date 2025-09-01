package files

import (
	"fmt"
	"sort"
	"strings"

	"ctxengine/internal/entity"
)

// PrintFileStruct выводит структуру FileCodeStruct в виде дерева.
func PrintFileStruct(fs *entity.FileCodeStruct, indent string) {
	fmt.Printf("%sFile: %s\n", indent, fs.FileName)
	if fs.Language != "" {
		fmt.Printf("%s  Language: %s\n", indent, fs.Language)
	}
	if len(fs.Keywords) > 0 {
		fmt.Printf("%s  Keywords: %v\n", indent, fs.Keywords)
	}
	if fs.PackageName != "" {
		fmt.Printf("%s  Package: %s\n", indent, fs.PackageName)
	}
	if fs.ModulePath != "" {
		fmt.Printf("%s  Module: %s\n", indent, fs.ModulePath)
	}
	if fs.Profile != nil {
		fmt.Printf("%s  Profile:\n", indent)
		fmt.Printf("%s    Name: %s\n", indent, fs.Profile.Name)
		if fs.Profile.TypeSystem != "" {
			fmt.Printf("%s    TypeSystem: %s\n", indent, fs.Profile.TypeSystem)
		}
		if len(fs.Profile.Paradigms) > 0 {
			fmt.Printf("%s    Paradigms: %v\n", indent, fs.Profile.Paradigms)
		}
		if fs.Profile.Inheritance != "" {
			fmt.Printf("%s    Inheritance: %s\n", indent, fs.Profile.Inheritance)
		}
		if fs.Profile.HasGenerics {
			fmt.Printf("%s    HasGenerics: %t\n", indent, fs.Profile.HasGenerics)
		}
		if fs.Profile.IsObjectOriented {
			fmt.Printf("%s    IsObjectOriented: %t\n", indent, fs.Profile.IsObjectOriented)
		}
		if fs.Profile.IsFunctional {
			fmt.Printf("%s    IsFunctional: %t\n", indent, fs.Profile.IsFunctional)
		}
	}

	// Imports
	if len(fs.Imports.System) > 0 {
		fmt.Printf("%s  Imports (System):\n", indent)
		for _, imp := range fs.Imports.System {
			fmt.Printf("%s    - %s\n", indent, imp)
		}
	}
	if len(fs.Imports.External) > 0 {
		fmt.Printf("%s  Imports (External):\n", indent)
		for _, imp := range fs.Imports.External {
			fmt.Printf("%s    - %s\n", indent, imp)
		}
	}
	if len(fs.Imports.Internal) > 0 {
		fmt.Printf("%s  Imports (Internal):\n", indent)
		for _, imp := range fs.Imports.Internal {
			fmt.Printf("%s    - %s\n", indent, imp)
		}
	}

	// Functions
	if len(fs.Functions) > 0 {
		fmt.Printf("%s  Functions:\n", indent)
		for _, f := range fs.Functions {
			printFuncInfo(f, indent+"    ")
		}
	}

	// Variables
	if len(fs.Variables) > 0 {
		fmt.Printf("%s  Variables:\n", indent)
		for _, v := range fs.Variables {
			fmt.Printf("%s    - Lines: %d-%d\n", indent, v.StartLine, v.EndLine)
		}
	}

	// Structs
	if len(fs.Structs) > 0 {
		fmt.Printf("%s  Structs:\n", indent)
		for _, s := range fs.Structs {
			printStructInfo(s, indent+"    ")
		}
	}

	// Classes
	if len(fs.Classes) > 0 {
		fmt.Printf("%s  Classes:\n", indent)
		for _, c := range fs.Classes {
			printClassInfo(c, indent+"    ")
		}
	}

	// Interfaces
	if len(fs.Interfaces) > 0 {
		fmt.Printf("%s  Interfaces:\n", indent)
		for _, i := range fs.Interfaces {
			printInterfaceInfo(i, indent+"    ")
		}
	}

	// Constants
	if len(fs.Constants) > 0 {
		fmt.Printf("%s  Constants:\n", indent)
		for _, c := range fs.Constants {
			fmt.Printf("%s    - Value: %s (Lines: %d-%d)\n", indent, c.Value, c.StartLine, c.EndLine)
		}
	}
}

// PrintAllIndexes печатает содержимое всех глобальных индексов: кода, функций,
// текстовых, YAML и Protobuf файлов. Предназначено для отладки и инспекции.
func PrintAllIndexes(code map[string]*entity.FileCodeStruct, codeFuncs map[string]*entity.FuncInfo,
	texts map[string]*entity.FileTextStruct, yamls map[string]*entity.FileYamlStruct,
	protos map[string]*entity.FileProtoStruct,
) {
	fmt.Println("==== GlobalCodeIndex ====")
	if len(code) == 0 {
		fmt.Println("<empty>")
	}
	for _, fs := range code {
		// Раньше печаталось дважды: здесь и внутри PrintFileStruct. Оставляем только PrintFileStruct.
		PrintFileStruct(fs, "")
	}

	fmt.Println("==== GlobalCodeFunctionIndex ====")
	if len(codeFuncs) == 0 {
		fmt.Println("<empty>")
	}
	for key, f := range codeFuncs {
		fmt.Printf("Key: %s\n", key)
		printFuncInfo(f, "  ")
	}

	fmt.Println("==== GlobalTextIndex ====")
	if len(texts) == 0 {
		fmt.Println("<empty>")
	}
	for file, ft := range texts {
		fmt.Printf("File: %s\n", file)
		// Чтобы не засорять вывод, печатаем только первые 200 символов
		preview := ft.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("  Content: %q\n", preview)
	}

	fmt.Println("==== GlobalYamlIndex ====")
	if len(yamls) == 0 {
		fmt.Println("<empty>")
	}
	for file, fy := range yamls {
		fmt.Printf("File: %s\n", file)
		preview := fy.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("  Content: %q\n", preview)
	}

	fmt.Println("==== GlobalProtoIndex ====")
	if len(protos) == 0 {
		fmt.Println("<empty>")
	}
	for file, fp := range protos {
		fmt.Printf("File: %s\n", file)
		preview := fp.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		fmt.Printf("  Content: %q\n", preview)
	}
}

// RenderAllIndexes возвращает строковое представление всех индексов
func RenderAllIndexes(code map[string]*entity.FileCodeStruct, codeFuncs map[string]*entity.FuncInfo,
	texts map[string]*entity.FileTextStruct, yamls map[string]*entity.FileYamlStruct,
	protos map[string]*entity.FileProtoStruct,
) string {
	var b strings.Builder
	b.WriteString("==== GlobalCodeIndex ====\n")
	if len(code) == 0 {
		b.WriteString("<empty>\n")
	}
	var files []string
	for f := range code {
		files = append(files, f)
	}
	sort.Strings(files)
	for _, fn := range files {
		fs := code[fn]
		b.WriteString(renderFileStruct(fs, ""))
	}

	b.WriteString("==== GlobalCodeFunctionIndex ====\n")
	if len(codeFuncs) == 0 {
		b.WriteString("<empty>\n")
	}
	var keys []string
	for k := range codeFuncs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		b.WriteString(fmt.Sprintf("Key: %s\n", k))
		b.WriteString(renderFuncInfo(codeFuncs[k], "  "))
	}

	b.WriteString("==== GlobalTextIndex ====\n")
	if len(texts) == 0 {
		b.WriteString("<empty>\n")
	}
	var textFiles []string
	for f := range texts {
		textFiles = append(textFiles, f)
	}
	sort.Strings(textFiles)
	for _, f := range textFiles {
		ft := texts[f]
		b.WriteString(fmt.Sprintf("File: %s\n", f))
		preview := ft.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		b.WriteString(fmt.Sprintf("  Content: %q\n", preview))
	}

	b.WriteString("==== GlobalYamlIndex ====\n")
	if len(yamls) == 0 {
		b.WriteString("<empty>\n")
	}
	var yamlFiles []string
	for f := range yamls {
		yamlFiles = append(yamlFiles, f)
	}
	sort.Strings(yamlFiles)
	for _, f := range yamlFiles {
		fy := yamls[f]
		b.WriteString(fmt.Sprintf("File: %s\n", f))
		preview := fy.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		b.WriteString(fmt.Sprintf("  Content: %q\n", preview))
	}

	b.WriteString("==== GlobalProtoIndex ====\n")
	if len(protos) == 0 {
		b.WriteString("<empty>\n")
	}
	var protoFiles []string
	for f := range protos {
		protoFiles = append(protoFiles, f)
	}
	sort.Strings(protoFiles)
	for _, f := range protoFiles {
		fp := protos[f]
		b.WriteString(fmt.Sprintf("File: %s\n", f))
		preview := fp.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		b.WriteString(fmt.Sprintf("  Content: %q\n", preview))
	}

	return b.String()
}

func renderFileStruct(fs *entity.FileCodeStruct, indent string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%sFile: %s\n", indent, fs.FileName))
	if fs.Language != "" {
		b.WriteString(fmt.Sprintf("%s  Language: %s\n", indent, fs.Language))
	}
	if len(fs.Keywords) > 0 {
		b.WriteString(fmt.Sprintf("%s  Keywords: %v\n", indent, fs.Keywords))
	}
	if fs.PackageName != "" {
		b.WriteString(fmt.Sprintf("%s  Package: %s\n", indent, fs.PackageName))
	}
	if fs.ModulePath != "" {
		b.WriteString(fmt.Sprintf("%s  Module: %s\n", indent, fs.ModulePath))
	}
	if fs.Profile != nil {
		b.WriteString(fmt.Sprintf("%s  Profile:\n", indent))
		b.WriteString(fmt.Sprintf("%s    Name: %s\n", indent, fs.Profile.Name))
		if fs.Profile.TypeSystem != "" {
			b.WriteString(fmt.Sprintf("%s    TypeSystem: %s\n", indent, fs.Profile.TypeSystem))
		}
		if len(fs.Profile.Paradigms) > 0 {
			b.WriteString(fmt.Sprintf("%s    Paradigms: %v\n", indent, fs.Profile.Paradigms))
		}
		if fs.Profile.Inheritance != "" {
			b.WriteString(fmt.Sprintf("%s    Inheritance: %s\n", indent, fs.Profile.Inheritance))
		}
		if fs.Profile.HasGenerics {
			b.WriteString(fmt.Sprintf("%s    HasGenerics: %t\n", indent, fs.Profile.HasGenerics))
		}
		if fs.Profile.IsObjectOriented {
			b.WriteString(fmt.Sprintf("%s    IsObjectOriented: %t\n", indent, fs.Profile.IsObjectOriented))
		}
		if fs.Profile.IsFunctional {
			b.WriteString(fmt.Sprintf("%s    IsFunctional: %t\n", indent, fs.Profile.IsFunctional))
		}
	}
	if len(fs.Imports.System) > 0 {
		b.WriteString(fmt.Sprintf("%s  Imports (System):\n", indent))
		for _, imp := range fs.Imports.System {
			b.WriteString(fmt.Sprintf("%s    - %s\n", indent, imp))
		}
	}
	if len(fs.Imports.External) > 0 {
		b.WriteString(fmt.Sprintf("%s  Imports (External):\n", indent))
		for _, imp := range fs.Imports.External {
			b.WriteString(fmt.Sprintf("%s    - %s\n", indent, imp))
		}
	}
	if len(fs.Imports.Internal) > 0 {
		b.WriteString(fmt.Sprintf("%s  Imports (Internal):\n", indent))
		for _, imp := range fs.Imports.Internal {
			b.WriteString(fmt.Sprintf("%s    - %s\n", indent, imp))
		}
	}
	if len(fs.Functions) > 0 {
		b.WriteString(fmt.Sprintf("%s  Functions:\n", indent))
		for _, f := range fs.Functions {
			b.WriteString(renderFuncInfo(f, indent+"    "))
		}
	}
	if len(fs.Variables) > 0 {
		b.WriteString(fmt.Sprintf("%s  Variables:\n", indent))
		for _, v := range fs.Variables {
			b.WriteString(fmt.Sprintf("%s    - Lines: %d-%d\n", indent, v.StartLine, v.EndLine))
		}
	}
	if len(fs.Structs) > 0 {
		b.WriteString(fmt.Sprintf("%s  Structs:\n", indent))
		for _, s := range fs.Structs {
			b.WriteString(renderStructInfo(s, indent+"    "))
		}
	}
	if len(fs.Classes) > 0 {
		b.WriteString(fmt.Sprintf("%s  Classes:\n", indent))
		for _, c := range fs.Classes {
			b.WriteString(renderClassInfo(c, indent+"    "))
		}
	}
	if len(fs.Interfaces) > 0 {
		b.WriteString(fmt.Sprintf("%s  Interfaces:\n", indent))
		for _, i := range fs.Interfaces {
			b.WriteString(renderInterfaceInfo(i, indent+"    "))
		}
	}
	if len(fs.Constants) > 0 {
		b.WriteString(fmt.Sprintf("%s  Constants:\n", indent))
		for _, c := range fs.Constants {
			b.WriteString(fmt.Sprintf("%s    - Value: %s (Lines: %d-%d)\n", indent, c.Value, c.StartLine, c.EndLine))
		}
	}
	return b.String()
}

func renderFuncInfo(f *entity.FuncInfo, indent string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s- Name: %s\n", indent, f.Name))
	if f.FQN != "" {
		b.WriteString(fmt.Sprintf("%s  FQN: %s\n", indent, f.FQN))
	}
	if f.OwnerName != "" {
		b.WriteString(fmt.Sprintf("%s  Owner: %s (%s)\n", indent, f.OwnerName, f.OwnerKind))
	}
	if f.ReceiverType != "" {
		b.WriteString(fmt.Sprintf("%s  ReceiverType: %s\n", indent, f.ReceiverType))
	}
	if f.IsConstructor {
		b.WriteString(fmt.Sprintf("%s  IsConstructor: %t\n", indent, f.IsConstructor))
	}
	if f.Access != "" {
		b.WriteString(fmt.Sprintf("%s  Access: %s\n", indent, f.Access))
	}
	if f.IsCall && f.CallKind != entity.CallKind("") {
		b.WriteString(fmt.Sprintf("%s  CallKind: %s\n", indent, f.CallKind))
	}
	if f.StartLine != 0 {
		b.WriteString(fmt.Sprintf("%s  StartLine: %d\n", indent, f.StartLine))
	}
	if f.EndLine != 0 {
		b.WriteString(fmt.Sprintf("%s  EndLine: %d\n", indent, f.EndLine))
	}
	if len(f.Comments) > 0 {
		b.WriteString(fmt.Sprintf("%s  Comments: %v\n", indent, f.Comments))
	}
	if len(f.Parameters) > 0 {
		b.WriteString(fmt.Sprintf("%s  Parameters:\n", indent))
		for _, p := range f.Parameters {
			if p.Value != "" {
				b.WriteString(fmt.Sprintf("%s    - Value: %s\n", indent, p.Value))
			} else {
				b.WriteString(fmt.Sprintf("%s    - Name: %s, Type: %s\n", indent, p.Name, p.Type))
			}
		}
	}
	if len(f.Returns) > 0 {
		b.WriteString(fmt.Sprintf("%s  Returns:\n", indent))
		for _, r := range f.Returns {
			if r.Value != "" {
				b.WriteString(fmt.Sprintf("%s    - Value: %s\n", indent, r.Value))
			} else {
				b.WriteString(fmt.Sprintf("%s    - Name: %s, Type: %s\n", indent, r.Name, r.Type))
			}
		}
	}
	if len(f.Functions) > 0 {
		b.WriteString(fmt.Sprintf("%s  Nested Functions:\n", indent))
		for _, nf := range f.Functions {
			b.WriteString(renderFuncInfo(nf, indent+"    "))
		}
	}
	if len(f.Overrides) > 0 {
		b.WriteString(fmt.Sprintf("%s  Overrides: %v\n", indent, f.Overrides))
	}
	if len(f.Implements) > 0 {
		b.WriteString(fmt.Sprintf("%s  Implements: %v\n", indent, f.Implements))
	}
	return b.String()
}

func renderStructInfo(s *entity.StructInfo, indent string) string {
	var b strings.Builder
	if s.Name != "" {
		b.WriteString(fmt.Sprintf("%s- Struct %s (Lines: %d-%d)\n", indent, s.Name, s.StartLine, s.EndLine))
	} else {
		b.WriteString(fmt.Sprintf("%s- Struct (Lines: %d-%d)\n", indent, s.StartLine, s.EndLine))
	}
	if s.FQN != "" {
		b.WriteString(fmt.Sprintf("%s  FQN: %s\n", indent, s.FQN))
	}
	if len(s.Comments) > 0 {
		b.WriteString(fmt.Sprintf("%s  Comments: %v\n", indent, s.Comments))
	}
	if len(s.Fields) > 0 {
		b.WriteString(fmt.Sprintf("%s  Fields:\n", indent))
		var names []string
		for n := range s.Fields {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, fieldName := range names {
			field := s.Fields[fieldName]
			b.WriteString(fmt.Sprintf("%s    - %s: %s (Lines: %d-%d)\n", indent, fieldName, field.Type, field.StartLine, field.EndLine))
		}
	}
	if len(s.Methods) > 0 {
		b.WriteString(fmt.Sprintf("%s  Methods:\n", indent))
		var names []string
		for n := range s.Methods {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			b.WriteString(fmt.Sprintf("%s    - %s:\n", indent, n))
			b.WriteString(renderFuncInfo(s.Methods[n], indent+"      "))
		}
	}
	if len(s.Implements) > 0 {
		b.WriteString(fmt.Sprintf("%s  Implements: %v\n", indent, s.Implements))
	}
	return b.String()
}

func renderClassInfo(c *entity.ClassInfo, indent string) string {
	var b strings.Builder
	name := c.Name
	if name == "" {
		name = "<anonymous>"
	}
	b.WriteString(fmt.Sprintf("%s- Class %s (Lines: %d-%d)\n", indent, name, c.StartLine, c.EndLine))
	if c.FQN != "" {
		b.WriteString(fmt.Sprintf("%s  FQN: %s\n", indent, c.FQN))
	}
	if len(c.Fields) > 0 {
		b.WriteString(fmt.Sprintf("%s  Fields:\n", indent))
		var names []string
		for n := range c.Fields {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			f := c.Fields[n]
			b.WriteString(fmt.Sprintf("%s    - %s: %s\n", indent, n, f.Type))
		}
	}
	if len(c.Methods) > 0 {
		b.WriteString(fmt.Sprintf("%s  Methods:\n", indent))
		var names []string
		for n := range c.Methods {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			b.WriteString(fmt.Sprintf("%s    - %s:\n", indent, n))
			b.WriteString(renderFuncInfo(c.Methods[n], indent+"      "))
		}
	}
	if len(c.Extends) > 0 {
		b.WriteString(fmt.Sprintf("%s  Extends: %v\n", indent, c.Extends))
	}
	if len(c.Implements) > 0 {
		b.WriteString(fmt.Sprintf("%s  Implements: %v\n", indent, c.Implements))
	}
	return b.String()
}

func renderInterfaceInfo(i *entity.InterfaceInfo, indent string) string {
	var b strings.Builder
	if i.Name != "" {
		b.WriteString(fmt.Sprintf("%s- Interface %s (Lines: %d-%d)\n", indent, i.Name, i.StartLine, i.EndLine))
	} else {
		b.WriteString(fmt.Sprintf("%s- Interface (Lines: %d-%d)\n", indent, i.StartLine, i.EndLine))
	}
	if i.FQN != "" {
		b.WriteString(fmt.Sprintf("%s  FQN: %s\n", indent, i.FQN))
	}
	if len(i.Methods) > 0 {
		b.WriteString(fmt.Sprintf("%s  Methods:\n", indent))
		var names []string
		for n := range i.Methods {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, name := range names {
			method := i.Methods[name]
			b.WriteString(fmt.Sprintf("%s    - %s:\n", indent, name))
			b.WriteString(renderFuncInfo(method, indent+"      "))
		}
	}
	if len(i.Extends) > 0 {
		b.WriteString(fmt.Sprintf("%s  Extends: %v\n", indent, i.Extends))
	}
	return b.String()
}

// printFuncInfo/printStructInfo/printClassInfo/printInterfaceInfo — обертки вокруг render* для печати
func printFuncInfo(f *entity.FuncInfo, indent string) {
	fmt.Print(renderFuncInfo(f, indent))
}

func printStructInfo(s *entity.StructInfo, indent string) {
	fmt.Print(renderStructInfo(s, indent))
}

func printClassInfo(c *entity.ClassInfo, indent string) {
	fmt.Print(renderClassInfo(c, indent))
}

func printInterfaceInfo(i *entity.InterfaceInfo, indent string) {
	fmt.Print(renderInterfaceInfo(i, indent))
}
