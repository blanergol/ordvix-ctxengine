package golang

import (
	"os"
	"path/filepath"
	"strings"
)

// DetectModulePath возвращает значение modulePath из файла go.mod в указанной директории проекта.
// В случае ошибки или отсутствия объявления module возвращает пустую строку.
func DetectModulePath(projectDir string) string {
	gomodPath := filepath.Join(projectDir, "go.mod")
	data, err := os.ReadFile(gomodPath)
	if err != nil {
		return ""
	}
	for _, l := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(l)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
