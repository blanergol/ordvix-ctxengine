package files

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"ctxengine/config"
	"ctxengine/internal/entity"
	"ctxengine/internal/golang"

	"github.com/fsnotify/fsnotify"
)

var (
	// GlobalCodeIndex Глобальный индекс файлов с кодом
	GlobalCodeIndex = make(map[string]*entity.FileCodeStruct)

	// GlobalCodeFunctionIndex Глобальный индекс функций кода
	GlobalCodeFunctionIndex = make(map[string]*entity.FuncInfo)

	// GlobalTextIndex Глобальный индекс файлов с текстом
	GlobalTextIndex = make(map[string]*entity.FileTextStruct)

	// GlobalYamlIndex Глобальный индекс файлов с текстом
	GlobalYamlIndex = make(map[string]*entity.FileYamlStruct)

	// GlobalProtoIndex Глобальный индекс файлов с текстом
	GlobalProtoIndex = make(map[string]*entity.FileProtoStruct)

	// FileModTimeCache хранит время последней модификации для каждого файла.
	FileModTimeCache = make(map[string]time.Time)
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
		// Ключ формируется по формату "FileName::FuncName"
		key := fmt.Sprintf("%s::%s", file, child.Name)
		if resolved, ok := GlobalCodeFunctionIndex[key]; ok {
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

func BuildGlobalFunctionIndex() {
	for file, fs := range GlobalCodeIndex {
		funcs := CollectFunctions(fs)
		for _, f := range funcs {
			if len(f.Name) == 0 {
				continue
			}
			key := fmt.Sprintf("%s::%s", file, f.Name)
			if _, exists := GlobalCodeFunctionIndex[key]; !exists {
				GlobalCodeFunctionIndex[key] = f
			}
		}
	}
}

// WatchAndReindex следит за изменениями в директории проекта, игнорируя .git
func WatchAndReindex(cfg *config.Config) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Ошибка создания наблюдателя: %v", err)
	}
	defer watcher.Close()

	// Рекурсивно добавляем наблюдение за директорией, игнорируя .git
	err = AddWatchersRecursively(watcher, cfg.Context.ProjectDir)
	if err != nil {
		log.Fatalf("Ошибка добавления наблюдения: %v", err)
	}

	log.Printf("Наблюдение запущено для директории: %s", cfg.Context.ProjectDir)

	// Дебаунс: если в течение N секунд не появилось новых событий, запускаем переиндексацию.
	debounceDelay := time.Duration(cfg.Context.TimeReIndex) * time.Second
	var timer *time.Timer

	resetTimer := func() {
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounceDelay, func() {
			log.Println("Нет изменений в течение 2 секунд. Запуск переиндексации...")
			ReindexWithWorkerPool(cfg)
		})
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Printf("Обнаружено событие: %s", event)
			resetTimer()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Ошибка наблюдения: %v", err)
		}
	}
}

// AddWatchersRecursively добавляет watcher для всех поддиректорий от корня,
// пропуская директории, имя которых содержит ".git".
func AddWatchersRecursively(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Если это директория и она содержит ".git" в имени, пропускаем её.
		if info.IsDir() {
			// Если имя директории равно ".git", пропускаем её и не идем вглубь.
			if strings.EqualFold(info.Name(), ".git") {
				return filepath.SkipDir
			}
			log.Printf("Добавляем наблюдение за директорией: %s", path)
			if err := watcher.Add(path); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReindexWithWorkerPool выполняет индексацию файлов с использованием пула воркеров.
func ReindexWithWorkerPool(cfg *config.Config) {
	log.Println("Запуск переиндексации с пулом воркеров...")

	allFiles, err := ReadDirectory(cfg.Context.ProjectDir)
	if err != nil {
		log.Printf("Ошибка обхода директории: %v", err)
		return
	}

	log.Printf("Найдено %d файлов", len(allFiles))
	startTime := time.Now()
	ctx := context.Background()

	// Локальные карты для результатов индексации
	newCodeIndex := make(map[string]*entity.FileCodeStruct)
	newCodeFuncIndex := make(map[string]*entity.FuncInfo)
	newTextIndex := make(map[string]*entity.FileTextStruct)
	newProtoIndex := make(map[string]*entity.FileProtoStruct)
	newYamlIndex := make(map[string]*entity.FileYamlStruct)

	var mu sync.Mutex
	var wg sync.WaitGroup

	workerPool := make(chan struct{}, 1000)

	for _, file := range allFiles {
		wg.Add(1)
		workerPool <- struct{}{} // блокируется, если достигнуто ограничение
		go func(file string) {
			defer wg.Done()
			defer func() { <-workerPool }()

			typeFile := DetectFileType(file)
			source, err := ParseFile(file)
			if err != nil {
				log.Printf("Ошибка чтения файла %s: %v", file, err)
				return
			}

			switch typeFile {
			case entity.TypeFileGolang:
				generated, err := golang.IsGeneratedFile(file, cfg.Context.ParseGenFile)
				if err != nil {
					log.Printf("Ошибка обработки файла %s: %v", file, err)
					return
				}

				fsStruct, err := golang.ParseFile(ctx, file, source, cfg.Context.ParseLevelDepth)
				if err != nil {
					log.Printf("Ошибка парсинга файла %s: %v", file, err)
					return
				}
				fsStruct.Generated = generated

				mu.Lock()
				newCodeIndex[file] = fsStruct
				// Собираем функции
				funcs := CollectFunctions(fsStruct)
				for _, f := range funcs {
					key := fmt.Sprintf("%s::%s", file, f.Name)
					newCodeFuncIndex[key] = f
				}
				mu.Unlock()
			case entity.TypeText:
				fsStruct := entity.FileTextStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newTextIndex[file] = &fsStruct
				mu.Unlock()
			case entity.TypeFileProtobuf:
				fsStruct := entity.FileProtoStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newProtoIndex[file] = &fsStruct
				mu.Unlock()
			case entity.TypeFileYaml:
				fsStruct := entity.FileYamlStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newYamlIndex[file] = &fsStruct
				mu.Unlock()
			default:
				log.Printf("Текущий язык программирования для файла %s не поддерживается", file)
			}
		}(file)
	}
	wg.Wait()

	// Обновляем глобальные индексы
	GlobalCodeIndex = newCodeIndex
	GlobalCodeFunctionIndex = newCodeFuncIndex
	GlobalTextIndex = newTextIndex
	GlobalProtoIndex = newProtoIndex
	GlobalYamlIndex = newYamlIndex

	// Построение глобального индекса функций и привязка дочерних функций
	BuildGlobalFunctionIndex()
	AttachAllFunctions(5)

	elapsed := time.Since(startTime)
	log.Printf("Переиндексация завершена за %s", elapsed)
}
