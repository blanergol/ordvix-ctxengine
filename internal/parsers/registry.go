package parsers

import (
	"context"

	"ctxengine/internal/entity"
)

// CodeParser описывает абстрактный парсер исходного кода для языка программирования.
// Парсер должен возвращать унифицированную структуру FileCodeStruct.
type CodeParser interface {
	Parse(ctx context.Context, fileName string, source []byte, maxDepth int, modulePath string) (*entity.FileCodeStruct, error)
}

var registry = map[string]CodeParser{}

// Register регистрирует парсер для указанного языка (ключ — тип файла из entity.TypeFile*).
func Register(lang string, parser CodeParser) {
	registry[lang] = parser
}

// Get возвращает зарегистрированный парсер по ключу языка.
func Get(lang string) (CodeParser, bool) {
	p, ok := registry[lang]
	return p, ok
}
