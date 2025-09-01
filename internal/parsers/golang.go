package parsers

import (
	"context"

	"ctxengine/internal/entity"
	"ctxengine/internal/golang"
)

// GolangParser адаптирует существующую реализацию парсинга Go к интерфейсу CodeParser.
type GolangParser struct{}

func (GolangParser) Parse(ctx context.Context, fileName string, source []byte, maxDepth int, modulePath string) (*entity.FileCodeStruct, error) {
	return golang.ParseFile(ctx, fileName, source, maxDepth, modulePath)
}
