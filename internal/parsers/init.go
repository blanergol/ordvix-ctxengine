package parsers

import "ctxengine/internal/entity"

// Инициализируем и регистрируем доступные парсеры.
func init() {
	Register(entity.TypeFileGolang, GolangParser{})
	Register(entity.TypeFilePython, PythonParser{})
}
