# CtxEngine

CtxEngine – это инструмент для анализа, индексирования и парсинга исходного кода (например, Go, Python, Java, TypeScript, Protobuf, YAML, Markdown, HTML и текстовых файлов). Инструмент может построить абстрактное синтаксическое дерево вызовов, позволяет выполнять быстрый поиск и анализ кода с возможностью повторной индексации при изменениях файлов.

## Функциональность

- **Парсинг файлов**  
  Индексирование исходного кода для языков, определённых в проекте (например, Go). Для остальных типов файлов (текст, YAML, Protobuf и т.д.) сохраняется их содержимое.

- **Построение глобальных индексов**  
  Сбор информации о файлах и функциях, построение глобальных карт для быстрого поиска по ключу формата `FQN` и `FileName::FuncName` (для обратной совместимости). Также строится индекс функций и выполняется привязка вызовов к определениям (с ограниченной глубиной).

- **Динамическое обновление индекса**  
  Отслеживание изменений в файловой системе с использованием [fsnotify](https://github.com/fsnotify/fsnotify). При обнаружении изменений с применением дебаунса (задержки) автоматически выполняется переиндексация.

- **Кэширование изменений**  
  Использование кэша времени модификации файлов для пропуска повторной обработки неизменённых файлов.

- **Параллельная обработка**  
  Использование пула воркеров для параллельного чтения и парсинга файлов, что позволяет ускорить индексирование при большом количестве файлов.

- **Языковой профиль и универсальная схема**  
  Единая модель представления кода для разных языков: с профилем языка (типизация, парадигмы, наследование), поддержкой `FQN`, владельца методов, интерфейсов/структур/классов, а также сбором ключевых слов и комментариев с привязкой к сущностям кода.

## Поддерживаемые языки и конструкции

На данный момент наиболее полно поддерживаются Go и Python. Для остальных типов файлов сохраняется содержимое как текст/документ.

### Go (через tree-sitter-golang)
- Объявления функций и методов: `function_declaration`, `method_declaration`
- Вызовы функций: `call_expression`  
  Классификация вызовов (`CallKind`):
  - `builtin` — встроенные функции (`append`, `len`, и т.п.)
  - `function` — вызов пакетной функции (`pkg.Func` или `Func`)
  - `method` — вызов метода у значения/указателя (`obj.Method`)
- Импорты: `import_declaration` (с разделением System/External/Internal)
- Переменные: `var_declaration` (глобальные)
- Константы: `const_declaration` (глобальные)
- Типы: `type_declaration` (структуры и интерфейсы, грубый разбор полей и сигнатур)
- `FQN` для типов и функций (на основе `package`/владельца)
- Сбор ключевых слов Go в файле
- Привязка ведущих комментариев к сущностям (функции, типы, константы, переменные)

Ограничения: привязка вызовов к определениям выполняется с ограниченной глубиной; разбор тел методов и полей — упрощённый.

### Python (через tree-sitter-python)
- Определения функций: `function_definition`
- Вызовы: `call` (отличение `obj.method` и `func`)
- Определения классов: `class_definition` (грубое извлечение методов и простейших полей вида `self.x = ...`)
- Импорты: `import_statement`, `import_from_statement`
- Сбор ключевых слов Python в файле
- `FQN` для классов и методов planned (частично)

Ограничения: извлечение полей класса — эвристическое; аннотации типов и наследование — базовые.

## Результирующая структура индексации (универсальная)

Ниже описаны ключевые сущности из `internal/entity`.

### FileCodeStruct (на файл исходного кода)
- `FileName` — путь к файлу
- `Language` — язык (`golang`, `python`, `java`, `typescript`, `protobuf`, `yaml`, `markdown`, `html`, `text`)
- `Generated` — сгенерирован ли файл
- `Functions` — список функций/методов верхнего уровня
- `Variables` — глобальные переменные (диапазоны строк)
- `Structs` — структуры (Go)
- `Classes` — классы (Python и ОО-языки)
- `Interfaces` — интерфейсы (Go/ОО-языки)
- `Constants` — глобальные константы
- `Imports` — импортируемые пакеты (`System`/`External`/`Internal`)
- `Keywords` — набор ключевых слов, встреченных в файле
- `PackageName` — имя пакета/пространства имен
- `ModulePath` — путь модуля (например, из go.mod)
- `Profile` — языковой профиль:
  - `Name` — имя языка
  - `TypeSystem` — типизация (`static`/`dynamic`/`gradual`)
  - `Paradigms` — поддерживаемые парадигмы (`oop`, `functional`, `procedural`, `generic`, `declarative`)
  - `Inheritance` — модель наследования (`class`, `interface`, `prototype`)
  - `HasGenerics`, `IsObjectOriented`, `IsFunctional`

### FuncInfo (функции и вызовы)
- `TokenInfo` — `StartLine`, `EndLine`, `Comments` (ведущие комментарии)
- `FileName`, `Name`, `Content`
- `Parameters`, `Returns` — списки `ParamInfo` (имя, тип, значение для аргумента вызова)
- `Functions` — вложенные вызовы/функции (дерево вызовов)
- `IsCall` — признак «вызов» (а не определение)
- `CallKind` — `function` | `method` | `builtin`  
  Для Go `pkg.Func` → `function`, `obj.Method` → `method`, `len/append/...` → `builtin`.
- `FQN` — полное имя (`package[.owner]::name`)
- `OwnerName`, `OwnerKind` — владелец метода (класс/интерфейс/структура)
- `ReceiverType` — тип получателя (Go методы)
- `IsConstructor`, `Access` — при наличии в языке
- `Overrides`, `Implements` — переопределения и реализации интерфейсов

### StructInfo / ClassInfo / InterfaceInfo
- Общие поля: `TokenInfo` (строки, комментарии), `Name`, `FQN`
- `Fields` — поля/атрибуты (`FieldInfo`: `Name`, `Type`, `TokenInfo`)
- `Methods` — карта методов (`name` → `FuncInfo`)
- `Implements` — реализуемые интерфейсы/протоколы
- `Extends` — базовые классы (для `ClassInfo`)

### ImportBlock
- `System` — стандартные/системные импорты
- `External` — внешние пакеты
- `Internal` — импорты текущего модуля

## Требования

- Go 1.18+
- Модули Go
- [fsnotify](https://github.com/fsnotify/fsnotify)
- [go-tree-sitter](https://github.com/smacker/go-tree-sitter)

## Установка

1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/yourusername/ctxengine.git
   cd ctxengine
   ```
2.	Установите зависимости:
   ```bash
  go mod tidy
   ```

3. Конфигурация

Конфигурация проекта задается через переменные окружения. Вы можете создать файл .env или установить переменные напрямую.
Пример файла .env:
```bash
# Настройки Orchestrator
ORCHESTRATOR_HOST=https://orchestrator.example.com
ORCHESTRATOR_TOKEN=your_orchestrator_token

# Настройки Context
CONTEXT_PROJECT_DIR=/path/to/your/project
CONTEXT_PARSE_GEN_FILE=false
CONTEXT_SKIP_GENERATED=false
CONTEXT_PARSE_LEVEL_DEPTH=100
CONTEXT_WATCH_ENABLED=false

# Настройки Rag
RAG_HOST=https://rag.example.com
RAG_TOKEN=your_rag_token
```

В конфигурации используется библиотека viper для автоматического считывания переменных окружения. Обратите внимание, что ключ context.project_dir соответствует переменной окружения CONTEXT_PROJECT_DIR благодаря замене точек на нижнее подчеркивание.

### Пропуск сгенерированных файлов

По умолчанию индексируются все файлы. Чтобы пропускать сгенерированные Go-файлы с маркерами вида "Code generated ... DO NOT EDIT", установите флаг:
```bash
export CONTEXT_SKIP_GENERATED=true
```
При этом в индексе такие файлы будут полностью пропущены. Если флаг отключён (default), все файлы индексируются, а поле `Generated` внутри `FileCodeStruct` будет выставлено в `true` для помеченных файлов.

## Использование

### Запуск индексации

Основной исполняемый файл проекта находится в cmd/main.go. При запуске приложение:
1.	Считывает конфигурацию из переменных окружения.
2.	Получает список файлов из указанной директории проекта.
3.	Индексирует файлы параллельно, используя пул воркеров и кэширование времени модификации.
4.	Строит глобальные индексы для кода, текста, YAML и Protobuf файлов.
5.	Запускает горутину, которая отслеживает изменения в директории и выполняет переиндексацию через заданный дебаунс.

Запустите приложение:
```bash
go run cmd/main.go
```

Чтобы включить наблюдение за изменениями файлов (по умолчанию выключено) и удерживать процесс запущенным, установите:
```bash
export CONTEXT_WATCH_ENABLED=true
```
Если флаг не задан или равен `false`, приложение выполнит индексацию и завершится.

### Структура проекта
- cmd/ – содержит исполняемые файлы (например, main.go).
- config/ – модуль для работы с конфигурацией (использует viper).
- internal/entity/ – содержит определения структур данных для представления кода, функций, типов файлов и прочего.
- internal/files/ – реализует функции для обхода директорий, чтения файлов, индексирования и отслеживания изменений (с использованием fsnotify).
- internal/golang/ – содержит функции для парсинга Go-кода с использованием go-tree-sitter.

### Оптимизация и масштабирование
- Пул воркеров и параллелизм:
Использование канала для ограничения количества одновременных горутин помогает ускорить индексацию и избежать чрезмерного использования ресурсов.
- Кэширование изменений:
Проверка времени модификации файлов позволяет не переиндексировать неизменённые файлы, что значительно снижает нагрузку на систему.
- Дебаунс изменений:
Использование таймера для дебаунса событий fsnotify позволяет запускать переиндексацию только после того, как изменения прекратятся на заданный промежуток времени.

## Пример запуска

Минимальный пример:
```bash
export CONTEXT_PROJECT_DIR=/abs/path/to/your/project
go run cmd/main.go
```

С включённой печатью индекса и увеличенной глубиной разбора:
```bash
export CONTEXT_PROJECT_DIR=/abs/path/to/your/project
export CONTEXT_PRINT_INDEX=true
export CONTEXT_PARSE_LEVEL_DEPTH=200
go run cmd/main.go
```

## Отладка (debug)

- Включите `CONTEXT_PRINT_INDEX=true` — после индексации артефакт будет записан в `debug/result_{n}.txt`.
- Смотрите секции:
  - `==== GlobalCodeIndex ====` — детализированная структура по файлам
  - `==== GlobalCodeFunctionIndex ====` — хеш-таблица функций по ключам (включая `FQN`)
  - `==== GlobalTextIndex ====`, `Yaml`, `Proto` — текстовые индексы
- Для трассировки переразборов используйте логи и параметр `CONTEXT_TIME_REINDEX` — по истечении N секунд без событий запускается переиндексация.
- При большом количестве файлов можно понизить `CONTEXT_PARSE_LEVEL_DEPTH` и/или отключить разбор сгенерированных файлов `CONTEXT_PARSE_GEN_FILE=false`.

## Примеры интерпретации вывода

Фрагмент (Go):
```
Functions:
  - Name: main
    FQN: ::main
    StartLine: 15
    EndLine: 30
    Comments: ["// main — точка входа приложения: загружает конфигурацию, выполняет первичную индексацию, // запускает наблюдение за изменениями и ожидает сигнала завершения."]
    Nested Functions:
      - Name: config.NewConfig
        CallKind: function
```

Пояснения:
- `FQN` без пакета для функции `main` — топ-уровень текущего пакета.
- `config.NewConfig` классифицирован как `function`, т.к. это вызов пакетной функции, а не метода объекта.

## Поддерживаемые операторы/конструкции (сводно)

- Go: объявления/вызовы функций и методов, импорты, var/const, структуры, интерфейсы, ключевые слова.
- Python: функции, вызовы, классы (методы и простые поля), импорты, ключевые слова.
- Прочие: YAML/Proto/Markdown/HTML/Text — сохраняются как контент.

Если нужно расширить парсинг (например, операторы выражений, блоки `try/except`, match/switch с ветками), добавляйте правила в соответствующие парсеры tree-sitter и в обработчики узлов.

## Развёртывание через Docker / Docker Compose

Минимальный запуск через Docker Compose. Файл `docker-compose.yml` уже присутствует в репозитории.

1. Экспортируйте переменные окружения (минимум — путь к проекту):
   ```bash
   export PROJECT_DIR=/abs/path/to/your/project
   export CONTEXT_WATCH_ENABLED=true         # опционально: следить за изменениями
   export CONTEXT_PRINT_INDEX=true           # опционально: сохранять отладочный дамп индексов
   ```
2. Запустите:
   ```bash
   docker compose up --build
   ```

Примечания:
- В контейнер пробрасываются тома:
  - `.:/app` — исходники `ctxengine`;
  - `${PROJECT_DIR}:/host_project:ro` — индексируемый проект (только чтение). Внутри контейнера `CONTEXT_PROJECT_DIR` установлен как `/host_project`.
- Команда запуска внутри контейнера: `go build -o /app/bin/ctxengine ./cmd && /app/bin/ctxengine`.
- Для длительного фонового наблюдения установите `CONTEXT_WATCH_ENABLED=true`.
- Отладочный вывод индексов при `CONTEXT_PRINT_INDEX=true` будет записан в `/app/debug/result_{n}.txt` внутри контейнера.

## Сборка и запуск бинаря (локально)

```bash
go build -o bin/ctxengine ./cmd
./bin/ctxengine
```

Рекомендуемые инструменты разработки:
- Makefile предоставляет цели:
  - `make fmt` — форматирование кода (`gofumpt`);
  - `make imports` — упорядочивание импортов (`goimports`);
  - `make test` — запуск юнит-тестов;
  - `make help` — список целей.

Установите gofumpt и goimports при необходимости:
```bash
go install mvdan.cc/gofumpt@latest
go install golang.org/x/tools/cmd/goimports@latest
```

## Переменные окружения

Все ключи читаются через `viper`, точки в ключах заменены на подчёркивания: `context.project_dir` → `CONTEXT_PROJECT_DIR`.

- Orchestrator
  - `ORCHESTRATOR_HOST` — базовый URL оркестратора (опционально).
  - `ORCHESTRATOR_TOKEN` — токен доступа (опционально).

- Context
  - `CONTEXT_PROJECT_DIR` — абсолютный путь к индексируемому проекту. Обязателен.
  - `CONTEXT_PARSE_GEN_FILE` — зарезервировано (парсить генерируемые файлы как обычные). По умолчанию `false`.
  - `CONTEXT_SKIP_GENERATED` — пропускать сгенерированные файлы (по сигнатурам «Code generated… DO NOT EDIT» и др.). По умолчанию `false`.
  - `CONTEXT_PARSE_LEVEL_DEPTH` — максимальная глубина обхода AST. По умолчанию `100`.
  - `CONTEXT_TIME_REINDEX` — дебаунс (секунды) перед переиндексацией при изменениях. По умолчанию `2`.
  - `CONTEXT_PRINT_INDEX` — сохранять отладочный дамп индексов в `debug/result_{n}.txt`. По умолчанию `false`.
  - `CONTEXT_WATCH_ENABLED` — включить наблюдение и переиндексацию по событиям fsnotify. По умолчанию `false`.

- Rag
  - `RAG_HOST` — базовый URL RAG-сервиса (опционально).
  - `RAG_TOKEN` — токен доступа (опционально).

## Программное использование (Go)

Простейший пример запуска индексации из вашего кода:

```go
package main

import (
  "log"
  "ctxengine/config"
  "ctxengine/internal/files"
)

func main() {
  cfg, err := config.NewConfig()
  if err != nil { log.Fatal(err) }

  // Одноразовая индексация
  files.ReindexWithWorkerPool(cfg)

  // Доступ к глобальным индексам
  for path, fs := range files.GlobalCodeIndex {
    log.Printf("Indexed file: %s, lang=%s, funcs=%d", path, fs.Language, len(fs.Functions))
  }
}
```

Для режима «наблюдать изменения»:

```go
cfg.Context.WatchEnabled = true
go files.WatchAndReindex(cfg)
// Заблокируйтесь любым удобным способом (канал, сигнал и т.д.)
select {}
```

Вывод индексов в текстовом виде (для логов/отладки):

```go
content := files.RenderAllIndexes(
  files.GlobalCodeIndex,
  files.GlobalCodeFunctionIndex,
  files.GlobalTextIndex,
  files.GlobalYamlIndex,
  files.GlobalProtoIndex,
)
log.Print(content)
```

## Как устроено определение типа файла

Функция `internal/files/files.go:DetectFileType` использует сочетание расширений и лёгких эвристик содержимого. Ключевые правила:
- `.go`, `.py`, `.java`, `.ts`/`.tsx`, `.proto`, `.yaml`/`.yml`, `.md`, `.html`/`.htm` — маппятся напрямую;
- Если расширение не распознано — читаются первые ~2 КБ и проверяются признаки:
  - `<!doctype html>`/`<html` → HTML;
  - префикс `---` → YAML;
  - наличие `package ...` и `class` → Java;
  - префикс `package` → Go;
  - shebang `#!/...python` → Python;
  - `import ... from ...` → TypeScript;
  - `syntax` и `proto3` → Protobuf;
- иначе — `text`.

## Расширение: добавление нового парсера

Чтобы поддержать новый язык:
1. Реализуйте `internal/parsers/<lang>.go`, соответствующий интерфейсу:
   ```go
   type CodeParser interface {
     Parse(ctx context.Context, fileName string, source []byte, maxDepth int, modulePath string) (*entity.FileCodeStruct, error)
   }
   ```
2. Зарегистрируйте парсер в `internal/parsers/init.go`:
   ```go
   func init() {
     parsers.Register(entity.TypeFileYourLang, YourLangParser{})
   }
   ```
3. Обновите `DetectFileType` для распознавания расширений/эвристик вашего языка.

Рекомендации:
- Наполняйте `FileCodeStruct.Profile` адекватными характеристиками языка.
- Для классов/структур/интерфейсов формируйте `FQN` и старайтесь связывать комментарии через «leading comments».
- Если возможно, классифицируйте вызовы на `function`/`method`/`builtin` для дальнейшей привязки.

## Поиск и разрешение вызовов

После индексации:
- `GlobalCodeIndex` — карта `file → FileCodeStruct`;
- `GlobalCodeFunctionIndex` — ключи: `FQN` и `"<file>::<funcName>"` (для обратной совместимости);
- `GlobalTextIndex`, `GlobalYamlIndex`, `GlobalProtoIndex` — текстовые индексы.

Пример извлечения функции по ключу:

```go
if f, ok := files.GlobalCodeFunctionIndex["main::init"]; ok {
  // ... используйте f (тип *entity.FuncInfo)
}
```

Привязка вызовов к определениям выполняется функцией `AttachAllFunctions(maxDepth)`. Алгоритм:
1. Строится `GlobalCodeFunctionIndex` по всем определённым функциям (`BuildGlobalFunctionIndex`).
2. Для каждого файла рекурсивно обходятся вложенные вызовы; если дочерний элемент — вызов пользовательской функции (`CallKind=function`), совершается попытка замены на определение по ключам `FQN` → `file::name`.
3. Глубина ограничена `maxDepth`.

Тонкости:
- Методы (`CallKind=method`) и `builtin` не перепривязываются.
- Для Go `FQN` формируется как `package[.owner]::name`.

## Makefile и локальная разработка

Полезные команды:
```bash
make fmt       # форматирование gofumpt
make imports   # упорядочивание импортов
make test      # юнит-тесты
```

Структура проекта (кратко):
- `cmd/` — точка входа (`main.go`);
- `config/` — конфигурация (viper);
- `internal/entity/` — модель домена и унифицированные структуры индекса;
- `internal/files/` — обход директорий, детект типов, индексация, наблюдение, форматирование вывода;
- `internal/golang/`, `internal/python/` — парсеры через tree-sitter;
- `internal/parsers/` — реестр и адаптеры парсеров.

## Типичные ошибки и диагностика

- «Пустой индекс/нет файлов» — проверьте `CONTEXT_PROJECT_DIR` и доступность пути (в Docker том должен существовать, флаг `:ro` делает его только для чтения, но этого достаточно).
- «go.mod не найден» — `modulePath` будет пустым, импорты `Internal` не будут определяться; это не критично, но влияет на классификацию импортов и формирование `FQN`.
- «Чрезмерная нагрузка при больших репозиториях» — уменьшите `CONTEXT_PARSE_LEVEL_DEPTH`, включите `CONTEXT_SKIP_GENERATED`, отключите `CONTEXT_PRINT_INDEX`, используйте `CONTEXT_WATCH_ENABLED=false` для одноразовой индексации.
- «Переиндексация не срабатывает» — проверьте `CONTEXT_WATCH_ENABLED=true` и `CONTEXT_TIME_REINDEX`. Убедитесь, что изменения происходят внутри наблюдаемой директории; `.git` игнорируется.
- «Файл определён как text, а не язык» — добавьте/уточните эвристику в `DetectFileType` или убедитесь в корректном расширении файла.

## FAQ

- Как найти все вызовы определённой функции?
  - Постройте `GlobalCodeFunctionIndex` (он строится автоматически) и пройдитесь по всем `FileCodeStruct.Functions`, собирая вложенные элементы с `IsCall=true` и совпадающим `Name`/`FQN`.

- Как изменить количество воркеров?
  - Сейчас `workerPoolSize` захардкожен (см. `internal/files/index.go`). При необходимости параметризуйте константу.

- Можно ли индексировать без доступа на запись к проекту?
  - Да. Индексация читает файлы; запись нужна только для отладочных дампов в `debug/` при `CONTEXT_PRINT_INDEX=true`.

- Как добавить поддержку JavaScript/TypeScript глубже текстового уровня?
  - Реализуйте парсер на базе `tree-sitter` (см. раздел «Расширение: добавление парсера») и зарегистрируйте его в `parsers.init`.
