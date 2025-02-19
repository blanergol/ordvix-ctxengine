package files

import (
	"ctxengine/internal/entity"
	"fmt"
)

// PrintFileStruct выводит структуру FileCodeStruct в виде дерева.
func PrintFileStruct(fs *entity.FileCodeStruct, indent string) {
	fmt.Printf("%sFile: %s\n", indent, fs.FileName)

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

func printFuncInfo(f *entity.FuncInfo, indent string) {
	fmt.Printf("%s- Name: %s\n", indent, f.Name)
	if f.StartLine != 0 {
		fmt.Printf("%s  StartLine: %d\n", indent, f.StartLine)
	}
	if f.EndLine != 0 {
		fmt.Printf("%s  EndLine: %d\n", indent, f.EndLine)
	}

	// Parameters
	if len(f.Parameters) > 0 {
		fmt.Printf("%s  Parameters:\n", indent)
		for _, p := range f.Parameters {
			fmt.Printf("%s    - Name: %s, Type: %s\n", indent, p.Name, p.Type)
		}
	}

	// Returns
	if len(f.Returns) > 0 {
		fmt.Printf("%s  Returns:\n", indent)
		for _, r := range f.Returns {
			fmt.Printf("%s    - Name: %s, Type: %s\n", indent, r.Name, r.Type)
		}
	}

	//// Content
	//if f.Content != "" {
	//	fmt.Printf("%s  Content:\n%s\n", indent, f.Content)
	//}

	// Nested Functions
	if len(f.Functions) > 0 {
		fmt.Printf("%s  Nested Functions:\n", indent)
		for _, nf := range f.Functions {
			printFuncInfo(nf, indent+"    ")
		}
	}
}

func printStructInfo(s *entity.StructInfo, indent string) {
	fmt.Printf("%s- Struct (Lines: %d-%d)\n", indent, s.StartLine, s.EndLine)
	if len(s.Comments) > 0 {
		fmt.Printf("%s  Comments: %v\n", indent, s.Comments)
	}
	if len(s.Fields) > 0 {
		fmt.Printf("%s  Fields:\n", indent)
		for fieldName, field := range s.Fields {
			fmt.Printf("%s    - %s: %s (Lines: %d-%d)\n", indent, fieldName, field.Type, field.StartLine, field.EndLine)
		}
	}
}

func printInterfaceInfo(i *entity.InterfaceInfo, indent string) {
	fmt.Printf("%s- Interface (Lines: %d-%d)\n", indent, i.StartLine, i.EndLine)
	if len(i.Methods) > 0 {
		fmt.Printf("%s  Methods:\n", indent)
		for methodName, method := range i.Methods {
			fmt.Printf("%s    - %s:\n", indent, methodName)
			printFuncInfo(method, indent+"      ")
		}
	}
}
