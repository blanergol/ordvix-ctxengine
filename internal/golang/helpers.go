package golang

import (
	"bufio"
	"os"
	"regexp"
)

// IsGeneratedFile проверяет, является ли файл сгенерированным, анализируя первые 10 строк на наличие типового комментария.
// Параметр check определяет, нужно ли проводить проверку (если false — сразу возвращается false).
func IsGeneratedFile(path string, check bool) (bool, error) {
	// Если проверка не требуется, сразу возвращаем false.
	if !check {
		return false, nil
	}

	// Открываем файл для чтения.
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Создаем сканер для чтения файла построчно.
	scanner := bufio.NewScanner(file)
	// Ограничиваем проверку первыми 10 строками файла.
	linesToCheck := 10

	// Регулярное выражение для поиска комментария, указывающего на то, что файл сгенерирован.
	// Оно ищет строки, начинающиеся с комментария (// или /*) и содержащие фразы "code generated" и "do not edit" (без учета регистра).
	generatedRegexp := regexp.MustCompile(`(?i)^(//|/\*)\s*code generated.*do not edit.*`)

	// Проходим по первым 10 строкам файла.
	for i := 0; i < linesToCheck && scanner.Scan(); i++ {
		line := scanner.Text()
		// Если регулярное выражение находит совпадение, файл считается сгенерированным.
		if generatedRegexp.MatchString(line) {
			return true, nil
		}
	}

	// Если во время сканирования произошла ошибка, возвращаем её.
	if err := scanner.Err(); err != nil {
		return false, err
	}

	// Если ни в одной из строк не найден типовой комментарий, файл не считается сгенерированным.
	return false, nil
}
