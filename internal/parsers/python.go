package parsers

import (
	"context"

	"ctxengine/internal/entity"
	"ctxengine/internal/python"
)

// PythonParser — реализация парсера Python через tree-sitter.
type PythonParser struct{}

func (PythonParser) Parse(ctx context.Context, fileName string, source []byte, maxDepth int, modulePath string) (*entity.FileCodeStruct, error) {
	return python.ParseFile(ctx, fileName, source, maxDepth, modulePath)
}
