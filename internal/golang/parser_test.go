package golang

import (
	"context"
	"testing"
)

func TestParseFileBasic(t *testing.T) {
	src := []byte("package p\nimport \"fmt\"\nfunc Hello(x int) string { fmt.Println(x); return \"ok\" }\n")
	fs, err := ParseFile(context.Background(), "hello.go", src, 10, "example.com/mod")
	if err != nil {
		t.Fatalf("ParseFile error: %v", err)
	}
	if fs.FileName != "hello.go" {
		t.Fatalf("filename not set")
	}
	if len(fs.Functions) == 0 {
		t.Fatalf("function Hello not parsed")
	}
	found := false
	for _, f := range fs.Functions {
		if f.Name == "Hello" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("function Hello not found")
	}
	if len(fs.Imports.System) == 0 { // fmt is stdlib (no dot)
		t.Fatalf("import not parsed")
	}
}
