package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadDirectory_SkipsGitAndBin(t *testing.T) {
	root := t.TempDir()
	// create .git and bin with files
	_ = os.MkdirAll(filepath.Join(root, ".git"), 0o755)
	_ = os.WriteFile(filepath.Join(root, ".git", "config"), []byte("dummy"), 0o600)
	_ = os.MkdirAll(filepath.Join(root, "bin"), 0o755)
	_ = os.WriteFile(filepath.Join(root, "bin", "tool"), []byte("dummy"), 0o700)
	// create regular file
	_ = os.MkdirAll(filepath.Join(root, "src"), 0o755)
	good := filepath.Join(root, "src", "file.txt")
	_ = os.WriteFile(good, []byte("ok"), 0o600)

	files, err := ReadDirectory(root)
	if err != nil {
		t.Fatalf("ReadDirectory error: %v", err)
	}
	for _, f := range files {
		if filepath.Base(filepath.Dir(f)) == ".git" || filepath.Base(filepath.Dir(f)) == "bin" {
			t.Fatalf("expected to skip .git and bin, but got: %s", f)
		}
	}
	// ensure our regular file is present
	found := false
	for _, f := range files {
		if f == good {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find %s", good)
	}
}
