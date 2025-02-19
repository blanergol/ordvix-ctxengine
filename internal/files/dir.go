package files

import (
	"io/fs"
	"path/filepath"
)

// ReadDirectory обходит указанную директорию (root) рекурсивно и возвращает список
// всех файлов, найденных в этой директории и во всех её поддиректориях.
// Если при обходе возникает ошибка, она возвращается сразу.
func ReadDirectory(root string) ([]string, error) {
	// Объявляем срез для хранения путей файлов
	var files []string

	// filepath.WalkDir обходит директорию рекурсивно.
	// Функция-обработчик вызывается для каждого файла и директории.
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// Если при обходе произошла ошибка, возвращаем её.
		if err != nil {
			return err
		}
		// Если текущий элемент не является директорией (то есть это файл),
		// добавляем его путь в срез.
		if !d.IsDir() {
			files = append(files, path)
		}
		// Возвращаем nil, чтобы продолжить обход.
		return nil
	})
	// Если ошибка произошла в процессе обхода, возвращаем её.
	if err != nil {
		return nil, err
	}
	// Возвращаем список файлов.
	return files, nil
}
