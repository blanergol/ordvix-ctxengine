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

	"ctxengine/internal/golang"

	"ctxengine/config"
	"ctxengine/internal/entity"
	"ctxengine/internal/parsers"
	"ctxengine/internal/utils"
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

const workerPoolSize = 100

// watcher/attach функции вынесены в watcher.go и attach.go

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

	// Определяем modulePath из go.mod один раз за переиндексацию, только если проект на Go
	modulePath := ""
	hasGo := false
	for _, f := range allFiles {
		if strings.HasSuffix(strings.ToLower(f), ".go") {
			hasGo = true
			break
		}
	}
	if hasGo {
		if _, err := os.Stat(filepath.Join(cfg.Context.ProjectDir, "go.mod")); err == nil {
			modulePath = golang.DetectModulePath(cfg.Context.ProjectDir)
		}
	}

	// Локальные карты для результатов индексации
	newCodeIndex := make(map[string]*entity.FileCodeStruct)
	newCodeFuncIndex := make(map[string]*entity.FuncInfo)
	newTextIndex := make(map[string]*entity.FileTextStruct)
	newProtoIndex := make(map[string]*entity.FileProtoStruct)
	newYamlIndex := make(map[string]*entity.FileYamlStruct)
	newModTimeCache := make(map[string]time.Time)

	var mu sync.Mutex
	var wg sync.WaitGroup

	workerPool := make(chan struct{}, workerPoolSize)

	for _, file := range allFiles {
		wg.Add(1)
		workerPool <- struct{}{} // блокируется, если достигнуто ограничение
		go func(file string) {
			defer wg.Done()
			defer func() { <-workerPool }()

			// Проверяем время модификации
			info, err := os.Stat(file)
			if err != nil {
				log.Printf("Не удалось получить Stat для %s: %v", file, err)
				return
			}
			modTime := info.ModTime()

			// Если файл не менялся, копируем из существующих индексов
			if prev, ok := FileModTimeCache[file]; ok && !modTime.After(prev) {
				mu.Lock()
				if fsStruct, ok := GlobalCodeIndex[file]; ok {
					newCodeIndex[file] = fsStruct
					// Переиндексируем функции для файла
					funcs := CollectFunctions(fsStruct)
					for _, f := range funcs {
						key := fmt.Sprintf("%s::%s", file, f.Name)
						newCodeFuncIndex[key] = f
					}
				} else if fsText, ok := GlobalTextIndex[file]; ok {
					newTextIndex[file] = fsText
				} else if fsProto, ok := GlobalProtoIndex[file]; ok {
					newProtoIndex[file] = fsProto
				} else if fsYaml, ok := GlobalYamlIndex[file]; ok {
					newYamlIndex[file] = fsYaml
				}
				newModTimeCache[file] = modTime
				mu.Unlock()
				return
			}

			typeFile := DetectFileType(file)
			source, err := ParseFile(file)
			if err != nil {
				log.Printf("Ошибка чтения файла %s: %v", file, err)
				return
			}

			switch typeFile {
			case entity.TypeFileGolang:
				// По умолчанию индексируем все файлы. Если включён флаг SkipGenerated — проверяем и при необходимости пропускаем
				generated := false
				if cfg.Context.SkipGenerated {
					var err error
					generated, err = utils.IsGeneratedFile(file, true)
					if err != nil {
						log.Printf("Ошибка обработки файла %s: %v", file, err)
						return
					}
					if generated {
						return
					}
				}
				parser, _ := parsers.Get(entity.TypeFileGolang)
				fsStruct, err := parser.Parse(ctx, file, source, cfg.Context.ParseLevelDepth, modulePath)
				if err != nil {
					log.Printf("Ошибка парсинга файла %s: %v", file, err)
					return
				}
				fsStruct.Generated = generated

				mu.Lock()
				newCodeIndex[file] = fsStruct
				funcs := CollectFunctions(fsStruct)
				for _, f := range funcs {
					key := fmt.Sprintf("%s::%s", file, f.Name)
					newCodeFuncIndex[key] = f
				}
				newModTimeCache[file] = modTime
				mu.Unlock()
			case entity.TypeFilePython:
				if parser, ok := parsers.Get(entity.TypeFilePython); ok {
					fsStruct, err := parser.Parse(ctx, file, source, cfg.Context.ParseLevelDepth, modulePath)
					if err != nil {
						log.Printf("Ошибка парсинга файла %s: %v", file, err)
						return
					}
					mu.Lock()
					newCodeIndex[file] = fsStruct
					funcs := CollectFunctions(fsStruct)
					for _, f := range funcs {
						key := fmt.Sprintf("%s::%s", file, f.Name)
						newCodeFuncIndex[key] = f
					}
					newModTimeCache[file] = modTime
					mu.Unlock()
				}
			case entity.TypeText:
				log.Printf("Текущий язык программирования для файла %s не поддерживается — сохраняем как текст", file)
				fsStruct := entity.FileTextStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newTextIndex[file] = &fsStruct
				newModTimeCache[file] = modTime
				mu.Unlock()
			case entity.TypeFileProtobuf:
				fsStruct := entity.FileProtoStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newProtoIndex[file] = &fsStruct
				newModTimeCache[file] = modTime
				mu.Unlock()
			case entity.TypeFileYaml:
				fsStruct := entity.FileYamlStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newYamlIndex[file] = &fsStruct
				newModTimeCache[file] = modTime
				mu.Unlock()
			default:
				// Для неподдерживаемых языков: логируем и сохраняем содержимое как текст
				log.Printf("Текущий язык программирования для файла %s не поддерживается — сохраняем как текст", file)
				fsStruct := entity.FileTextStruct{
					FileName: file,
					Content:  string(source),
				}
				mu.Lock()
				newTextIndex[file] = &fsStruct
				newModTimeCache[file] = modTime
				mu.Unlock()
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
	FileModTimeCache = newModTimeCache

	// Построение глобального индекса функций и привязка дочерних функций
	BuildGlobalFunctionIndex()
	AttachAllFunctions(5)

	elapsed := time.Since(startTime)
	log.Printf("Переиндексация завершена за %s", elapsed)
	if cfg.Context.PrintIndex {
		// Сохраняем результат индексации в файл debug/result_{n}.txt без вывода в консоль
		_ = os.MkdirAll("debug", 0o755)
		// Находим следующий инкремент
		n := 1
		for {
			candidate := filepath.Join("debug", fmt.Sprintf("result_%d.txt", n))
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				content := RenderAllIndexes(GlobalCodeIndex, GlobalCodeFunctionIndex, GlobalTextIndex, GlobalYamlIndex, GlobalProtoIndex)
				_ = os.WriteFile(candidate, []byte(content), 0o644)
				break
			}
			n++
		}
	}
}
