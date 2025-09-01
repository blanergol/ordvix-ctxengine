package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"ctxengine/internal/entity"
)

// DetectFileType определяет тип файла на основе расширения и содержимого.
func DetectFileType(fileName string) string {
	ext := strings.ToLower(filepath.Ext(fileName))
	switch ext {
	case ".go":
		return entity.TypeFileGolang
	case ".py":
		return entity.TypeFilePython
	case ".java":
		return entity.TypeFileJava
	case ".ts", ".tsx":
		return entity.TypeFileTypeScript
	case ".proto":
		return entity.TypeFileProtobuf
	case ".yaml", ".yml":
		return entity.TypeFileYaml
	case ".md":
		return entity.TypeMarkdown
	case ".html", ".htm":
		return entity.TypeHTML
	}

	// Если расширение не распознано, читаем первые 1024 байта файла.
	f, err := os.Open(fileName)
	if err != nil {
		return entity.TypeText
	}
	defer f.Close()

	buf := make([]byte, 2048)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return entity.TypeText
	}
	content := string(buf[:n])
	contentLower := strings.ToLower(content)

	// Проверяем наличие HTML-маркеров.
	if strings.Contains(contentLower, "<html") || strings.Contains(contentLower, "<!doctype html") {
		return entity.TypeHTML
	}

	// YAML: если файл начинается с "---".
	trimmed := strings.TrimSpace(content)
	if strings.HasPrefix(trimmed, "---") {
		return entity.TypeFileYaml
	}

	// Markdown: если встречаются заголовки вида "# " или "## ".
	if strings.Contains(content, "# ") || strings.Contains(content, "## ") {
		return entity.TypeMarkdown
	}

	// Java: если содержимое содержит и "package" и "class ".
	if strings.Contains(content, "package ") && strings.Contains(content, "class ") {
		return entity.TypeFileJava
	}

	// Go: если файл начинается с "package ".
	if strings.HasPrefix(trimmed, "package ") {
		return entity.TypeFileGolang
	}

	// Python: если первая строка — shebang с "python".
	if strings.HasPrefix(trimmed, "#!") && strings.Contains(contentLower, "python") {
		return entity.TypeFilePython
	}

	// TypeScript: если встречается конструкция "import ... from ".
	if strings.Contains(content, "import ") && strings.Contains(content, " from ") {
		return entity.TypeFileTypeScript
	}

	// Protobuf: если содержимое содержит "syntax" и "proto3".
	if strings.Contains(content, "syntax") && strings.Contains(content, "proto3") {
		return entity.TypeFileProtobuf
	}

	return entity.TypeText
}

// ParseFile читает файл по пути fileName и возвращает его содержимое как срез байт.
// В случае ошибки чтения возвращает непустую ошибку и пустой срез.
func ParseFile(fileName string) ([]byte, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return []byte{}, fmt.Errorf("не удалось прочитать файл %s: %v", fileName, err)
	}
	return data, nil
}
