package config

import (
	"os"
	"testing"
)

func TestNewConfigDefaults(t *testing.T) {
	t.Setenv("ORCHESTRATOR_HOST", "")
	t.Setenv("ORCHESTRATOR_TOKEN", "")
	t.Setenv("CONTEXT_PROJECT_DIR", "")
	t.Setenv("CONTEXT_PARSE_GEN_FILE", "false")
	t.Setenv("CONTEXT_PARSE_LEVEL_DEPTH", "")
	t.Setenv("CONTEXT_TIME_REINDEX", "")
	t.Setenv("RAG_HOST", "")
	t.Setenv("RAG_TOKEN", "")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig error: %v", err)
	}

	if cfg.Context.ParseLevelDepth != DefaultLevelDepth {
		t.Fatalf("expected default depth %d, got %d", DefaultLevelDepth, cfg.Context.ParseLevelDepth)
	}
	if cfg.Context.TimeReIndex != DefaultTimeReIndex {
		t.Fatalf("expected default time reindex %d, got %d", DefaultTimeReIndex, cfg.Context.TimeReIndex)
	}

	if cfg.Orchestrator.Host != "" || cfg.Rag.Host != "" {
		t.Fatalf("expected empty hosts by default")
	}

	_ = os.Unsetenv("CONTEXT_PROJECT_DIR")
}

func TestNewConfigFromEnv(t *testing.T) {
	t.Setenv("ORCHESTRATOR_HOST", "https://orch")
	t.Setenv("ORCHESTRATOR_TOKEN", "tok")
	t.Setenv("CONTEXT_PROJECT_DIR", "/tmp/x")
	t.Setenv("CONTEXT_PARSE_GEN_FILE", "true")
	t.Setenv("CONTEXT_PARSE_LEVEL_DEPTH", "77")
	t.Setenv("CONTEXT_TIME_REINDEX", "5")
	t.Setenv("RAG_HOST", "https://rag")
	t.Setenv("RAG_TOKEN", "rtok")

	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig error: %v", err)
	}

	if cfg.Orchestrator.Host != "https://orch" || cfg.Rag.Host != "https://rag" {
		t.Fatalf("hosts not loaded from env")
	}
	if cfg.Context.ProjectDir != "/tmp/x" || !cfg.Context.ParseGenFile {
		t.Fatalf("context not loaded from env")
	}
	if cfg.Context.ParseLevelDepth != 77 || cfg.Context.TimeReIndex != 5 {
		t.Fatalf("numeric context not loaded from env")
	}
}
