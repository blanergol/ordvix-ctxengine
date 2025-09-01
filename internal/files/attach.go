package files

import (
	"fmt"

	"ctxengine/internal/entity"
)

// attachFunctionsRecursively проходит по всем дочерним функциям родительской функции f
// и, если для дочерней функции найдено определение в GlobalCodeFunctionIndex, заменяет её.
// Глубина рекурсии регулируется параметром currentDepth, и не производится привязка, если currentDepth >= maxDepth.
func attachFunctionsRecursively(f *entity.FuncInfo, currentDepth, maxDepth int, file string) {
	if currentDepth >= maxDepth {
		return
	}

	// Сначала рекурсивно обходим все дочерние функции, чтобы привязать их вложенные вызовы.
	for _, child := range f.Functions {
		attachFunctionsRecursively(child, currentDepth+1, maxDepth, file)
	}

	// Затем для каждого дочернего элемента пытаемся найти глобальное определение по ключу.
	for i, child := range f.Functions {
		if len(child.Name) == 0 {
			continue
		}
		// Привязываем только пользовательские вызовы функций, игнорируем методы и builtin
		if child.IsCall && child.CallKind != entity.CallKindFunction {
			continue
		}
		// Ключи: сначала пробуем FQN, затем <file>::<name>
		if child.FQN != "" {
			if resolved, ok := GlobalCodeFunctionIndex[child.FQN]; ok {
				f.Functions[i] = resolved
				continue
			}
		}
		fallbackKey := fmt.Sprintf("%s::%s", file, child.Name)
		if resolved, ok := GlobalCodeFunctionIndex[fallbackKey]; ok {
			f.Functions[i] = resolved
		}
	}
}

// AttachAllFunctions Пример использования: для каждого файла в глобальном индексе
// запускаем привязку для всех верхнеуровневых функций.
func AttachAllFunctions(maxDepth int) {
	for file, fs := range GlobalCodeIndex {
		for i := range fs.Functions {
			attachFunctionsRecursively(fs.Functions[i], 1, maxDepth, file)
		}
	}
}

// CollectFunctions обходит структуру FileCodeStruct и рекурсивно собирает все FuncInfo.
func CollectFunctions(fs *entity.FileCodeStruct) []*entity.FuncInfo {
	var funcs []*entity.FuncInfo

	var traverse func(f *entity.FuncInfo)
	traverse = func(f *entity.FuncInfo) {
		funcs = append(funcs, f)
		for _, child := range f.Functions {
			traverse(child)
		}
	}

	for _, f := range fs.Functions {
		traverse(f)
	}
	return funcs
}

// BuildGlobalFunctionIndex строит глобальный индекс функций `GlobalCodeFunctionIndex`
// по ключу вида "<полный_путь_к_файлу>::<имя_функции>" для быстрых разрешений ссылок.
// Индекс используется на этапе привязки, чтобы подменять вызовы их фактическими определениями.
func BuildGlobalFunctionIndex() {
	for file, fs := range GlobalCodeIndex {
		funcs := CollectFunctions(fs)
		for _, f := range funcs {
			if len(f.Name) == 0 || f.IsCall {
				continue
			}
			// Индексируем по FQN, если есть
			if f.FQN != "" {
				if _, exists := GlobalCodeFunctionIndex[f.FQN]; !exists {
					GlobalCodeFunctionIndex[f.FQN] = f
				}
			}
			// И по старому ключу (на случай отсутствия информации о пакете)
			legacyKey := fmt.Sprintf("%s::%s", file, f.Name)
			if _, exists := GlobalCodeFunctionIndex[legacyKey]; !exists {
				GlobalCodeFunctionIndex[legacyKey] = f
			}
		}
	}
}
