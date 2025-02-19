package entity

// Константы, определяющие типы файлов, используемые в приложении.
const (
	// Файл написан на языке Go.
	TypeFileGolang = "golang"
	// Файл написан на языке Python.
	TypeFilePython = "python"
	// Файл написан на языке Java.
	TypeFileJava = "java"
	// Файл написан на языке TypeScript.
	TypeFileTypeScript = "typescript"
	// Файл является файлом protobuf.
	TypeFileProtobuf = "protobuf"
	// Файл является YAML-файлом.
	TypeFileYaml = "yaml"
	// Файл является Markdown документом.
	TypeMarkdown = "markdown"
	// Файл является HTML документом.
	TypeHTML = "html"
	// Файл считается текстовым (по умолчанию).
	TypeText = "text"
)

// TokenInfo хранит информацию о диапазоне строк и комментариях, связанных с элементом кода.
type TokenInfo struct {
	StartLine int      // Начальная строка элемента
	EndLine   int      // Конечная строка элемента
	Comments  []string // Связанные комментарии
}

// ParamInfo описывает параметр функции или метода.
type ParamInfo struct {
	Name string // Имя параметра
	Type string // Тип параметра
}

// FuncInfo представляет информацию об объявлении функции или метода.
type FuncInfo struct {
	TokenInfo               // Встраивание информации о строках и комментариях
	FileName   string       // Путь к файлу, где определена функция
	Name       string       // Имя функции
	Parameters []*ParamInfo // Список параметров функции
	Returns    []*ParamInfo // Список возвращаемых значений
	Functions  []*FuncInfo  // Вложенные функции или вызовы (например, анонимные функции)
	Content    string       // Исходный код функции (может использоваться для дальнейшего анализа)
}

// FieldInfo содержит информацию о поле структуры (например, имя и тип).
type FieldInfo struct {
	TokenInfo        // Диапазон строк и комментарии, связанные с полем
	Name      string // Имя поля
	Type      string // Тип поля
}

// StructInfo представляет структуру данных (структуру) в исходном коде.
type StructInfo struct {
	TokenInfo                       // Диапазон строк и комментарии
	Comments  []string              // Комментарии, относящиеся к структуре
	Fields    map[string]*FieldInfo // Поля структуры, индексированные по имени
}

// InterfaceInfo представляет информацию об интерфейсе, содержащем набор методов.
type InterfaceInfo struct {
	TokenInfo                      // Диапазон строк и комментарии
	Methods   map[string]*FuncInfo // Методы интерфейса, индексированные по имени
}

// ConstInfo описывает константу: диапазон строк и значение константы.
type ConstInfo struct {
	TokenInfo        // Диапазон строк и комментарии
	Value     string // Значение константы
}

// ImportBlock хранит информацию об импортируемых пакетах, разделяя системные и внешние импорты.
type ImportBlock struct {
	System   []string // Системные импорты (например, стандартная библиотека)
	External []string // Внешние импорты (например, сторонние библиотеки)
}

// FileCodeStruct представляет проиндексированный файл с кодом.
// В него включается информация о функциях, переменных, структурах, интерфейсах, константах и импортируемых пакетах.
type FileCodeStruct struct {
	FileName   string           // Путь к файлу
	Generated  bool             // Флаг: является ли файл сгенерированным
	Functions  []*FuncInfo      // Список функций в файле
	Variables  []*TokenInfo     // Переменные, объявленные вне функций
	Structs    []*StructInfo    // Структуры
	Interfaces []*InterfaceInfo // Интерфейсы
	Constants  []*ConstInfo     // Константы
	Imports    ImportBlock      // Импорты
}

// FileProtoStruct представляет проиндексированный файл с определением protobuf.
type FileProtoStruct struct {
	FileName string // Путь к файлу
	Gen      bool   // Флаг: является ли файл сгенерированным
	Content  string // Содержимое файла
}

// FileYamlStruct представляет проиндексированный YAML-файл.
type FileYamlStruct struct {
	FileName string // Путь к файлу
	Gen      bool   // Флаг: является ли файл сгенерированным
	Content  string // Содержимое файла
}

// FileTextStruct представляет проиндексированный текстовый файл.
type FileTextStruct struct {
	FileName string // Путь к файлу
	Gen      bool   // Флаг: является ли файл сгенерированным (если применимо)
	Content  string // Содержимое файла
}
