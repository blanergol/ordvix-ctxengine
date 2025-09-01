package files

import (
	"io/fs"
	"path/filepath"
)

// ReadDirectory рекурсивно обходит директорию root и возвращает список файлов (полные пути),
// пропуская служебные директории.
func ReadDirectory(root string) ([]string, error) {
	var files []string
	// WalkDir предпочтительнее Walk для работы с DirEntry
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Если при обходе произошла ошибка, возвращаем её.
		if err != nil {
			return err
		}
		// Пропускаем служебные директории целиком
		if d.IsDir() {
			switch d.Name() {
			case ".git", "bin", ".idea", ".vscode", ".vs", ".settings", ".fleet", "nbproject":
				return fs.SkipDir
			}
		}
		// Если текущий элемент не является директорией (то есть это файл),
		// добавляем его путь в срез.
		if !d.IsDir() {
			files = append(files, path)
		}
		// Возвращаем nil, чтобы продолжить обход.
		return nil
	})
	return files, err
}
