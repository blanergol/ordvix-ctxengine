package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectFileTypeByExt(t *testing.T) {
	cases := map[string]string{
		"a.go":       "golang",
		"b.py":       "python",
		"c.java":     "java",
		"d.ts":       "typescript",
		"e.proto":    "protobuf",
		"f.yaml":     "yaml",
		"g.md":       "markdown",
		"h.html":     "html",
		"unknown.xx": "text",
	}
	for name, want := range cases {
		got := DetectFileType(name)
		if got != want {
			t.Errorf("%s want %s got %s", name, want, got)
		}
	}
}

func TestDetectFileTypeByContent(t *testing.T) {
	dir := t.TempDir()

	// HTML
	html := filepath.Join(dir, "a.txt")
	os.WriteFile(html, []byte("<!doctype html><html><body></body></html>"), 0o600)
	if got := DetectFileType(html); got != "html" {
		t.Fatalf("html detect failed: %s", got)
	}

	// YAML
	yml := filepath.Join(dir, "b.txt")
	os.WriteFile(yml, []byte("---\nkey: value\n"), 0o600)
	if got := DetectFileType(yml); got != "yaml" {
		t.Fatalf("yaml detect failed: %s", got)
	}

	// Markdown
	md := filepath.Join(dir, "c.txt")
	os.WriteFile(md, []byte("# Header\ntext"), 0o600)
	if got := DetectFileType(md); got != "markdown" {
		t.Fatalf("markdown detect failed: %s", got)
	}
}

func TestReadDirectorySkipsGit(t *testing.T) {
	root := t.TempDir()
	os.Mkdir(filepath.Join(root, ".git"), 0o700)
	good := filepath.Join(root, "file.txt")
	os.WriteFile(good, []byte("x"), 0o600)

	files, err := ReadDirectory(root)
	if err != nil {
		t.Fatalf("ReadDirectory error: %v", err)
	}
	// Ensure only our file is returned
	if len(files) != 1 || files[0] != good {
		t.Fatalf("unexpected files: %#v", files)
	}
}
