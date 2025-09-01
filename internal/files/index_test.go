package files

import (
	"os"
	"path/filepath"
	"testing"

	"ctxengine/config"
)

// Проверяем, что неизвестные расширения попадают в текстовый индекс
func TestReindex_UnknownAsText(t *testing.T) {
	root := t.TempDir()
	// проект с одним .js файлом
	proj := filepath.Join(root, "proj")
	_ = os.MkdirAll(proj, 0o755)
	js := filepath.Join(proj, "x.js")
	_ = os.WriteFile(js, []byte("console.log('hi')\n"), 0o600)

	cfg := &config.Config{}
	cfg.Context.ProjectDir = proj
	cfg.Context.ParseLevelDepth = config.DefaultLevelDepth
	cfg.Context.PrintIndex = false

	ReindexWithWorkerPool(cfg)
	if _, ok := GlobalTextIndex[js]; !ok {
		t.Fatalf("expected %s to be indexed as text", js)
	}
}
