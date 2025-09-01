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

// Типизация языка программирования
type TypeSystemKind string

const (
	TypeSystemStatic  TypeSystemKind = "static"
	TypeSystemDynamic TypeSystemKind = "dynamic"
	TypeSystemGradual TypeSystemKind = "gradual"
)

// Парадигмы программирования
type ParadigmType string

const (
	ParadigmOOP         ParadigmType = "oop"
	ParadigmFunctional  ParadigmType = "functional"
	ParadigmProcedural  ParadigmType = "procedural"
	ParadigmGeneric     ParadigmType = "generic"
	ParadigmDeclarative ParadigmType = "declarative"
)

// Модель наследования
type InheritanceModel string

const (
	InheritanceClass     InheritanceModel = "class"
	InheritancePrototype InheritanceModel = "prototype"
	InheritanceInterface InheritanceModel = "interface"
)

// Профиль языка программирования (обобщенные характеристики)
type LanguageProfile struct {
	Name             string           // Имя языка (например, golang, python)
	TypeSystem       TypeSystemKind   // Типизация: static/dynamic/gradual
	Paradigms        []ParadigmType   // Поддерживаемые парадигмы
	Inheritance      InheritanceModel // Модель наследования
	HasGenerics      bool             // Поддерживает ли дженерики/параметрические типы
	IsObjectOriented bool             // Имеет ли полноценные классы/объекты
	IsFunctional     bool             // Имеет ли функциональные конструкции первого класса
}

// Константы для классификации вызовов функций
type CallKind string

const (
	CallKindFunction CallKind = "function" // Вызов пользовательской функции
	CallKindMethod   CallKind = "method"   // Вызов метода (selector_expression и аналоги)
	CallKindBuiltin  CallKind = "builtin"  // Встроенная функция языка
)

// TokenInfo хранит информацию о диапазоне строк и комментариях, связанных с элементом кода.
type TokenInfo struct {
	StartLine int      // Начальная строка элемента
	EndLine   int      // Конечная строка элемента
	Comments  []string // Связанные комментарии
}

// ParamInfo описывает параметр функции или метода.
type ParamInfo struct {
	Name  string // Имя параметра
	Type  string // Тип параметра
	Value string // Значение (для аргументов вызовов)
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
	IsCall     bool         // Позволяет отличать вызовы от объявлений
	CallKind   CallKind     // Классификация вызова: function/method/builtin

	// Полная квалификация и владение
	FQN           string   // Полное имя (пакет/модуль[.класс]::имя)
	OwnerName     string   // Владелец (класс/интерфейс/структура), если метод
	OwnerKind     string   // Тип владельца: class/interface/struct/module/package
	ReceiverType  string   // Тип получателя (например, *MyType в Go)
	IsConstructor bool     // Является ли конструктором/инициализатором
	Access        string   // Уровень доступа (public/protected/private и т.п.)
	Overrides     []string // FQN методов, которые переопределяет (наследование)
	Implements    []string // FQN методов интерфейсов, которые реализует
}

// FieldInfo содержит информацию о поле структуры (например, имя и тип).
type FieldInfo struct {
	TokenInfo        // Диапазон строк и комментарии, связанные с полем
	Name      string // Имя поля
	Type      string // Тип поля
}

// StructInfo представляет структуру данных (структуру) в исходном коде.
type StructInfo struct {
	TokenInfo                        // Диапазон строк и комментарии
	Name       string                // Имя структуры
	Comments   []string              // Комментарии, относящиеся к структуре
	Fields     map[string]*FieldInfo // Поля структуры, индексированные по имени
	Methods    map[string]*FuncInfo  // Методы, связанные с типом-получателем (как в Go)
	FQN        string                // Полное имя структуры (с пакетом/модулем)
	Implements []string              // Реализуемые интерфейсы (по совокупности методов)
}

// ClassInfo представляет класс (например, Python), содержащий атрибуты и методы.
type ClassInfo struct {
	TokenInfo                        // Диапазон строк и комментарии
	Name       string                // Имя класса
	Fields     map[string]*FieldInfo // Атрибуты/поля класса
	Methods    map[string]*FuncInfo  // Методы класса
	FQN        string                // Полное имя класса (с пакетом/модулем)
	Extends    []string              // Родительские классы (по FQN)
	Implements []string              // Реализуемые интерфейсы/протоколы (по FQN)
}

// InterfaceInfo представляет информацию об интерфейсе, содержащем набор методов.
type InterfaceInfo struct {
	TokenInfo                      // Диапазон строк и комментарии
	Name      string               // Имя интерфейса
	Methods   map[string]*FuncInfo // Методы интерфейса, индексированные по имени
	FQN       string               // Полное имя интерфейса
	Extends   []string             // Базовые интерфейсы (наследование интерфейсов)
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
	Internal []string // Импорты внутри текущего модуля/проекта
}

// FileCodeStruct представляет проиндексированный файл с кодом.
// В него включается информация о функциях, переменных, структурах, интерфейсах, константах и импортируемых пакетах.
type FileCodeStruct struct {
	FileName    string           // Путь к файлу
	Language    string           // Язык исходного файла (см. TypeFile*)
	Generated   bool             // Флаг: является ли файл сгенерированным
	Functions   []*FuncInfo      // Список функций в файле
	Variables   []*TokenInfo     // Переменные, объявленные вне функций
	Structs     []*StructInfo    // Структуры (например, Go)
	Classes     []*ClassInfo     // Классы (например, Python)
	Interfaces  []*InterfaceInfo // Интерфейсы
	Constants   []*ConstInfo     // Константы
	Imports     ImportBlock      // Импорты
	Keywords    []string         // Использованные в файле ключевые слова языка
	PackageName string           // Имя пакета/пространства имен (если применимо)
	ModulePath  string           // Путь/идентификатор модуля
	Profile     *LanguageProfile // Профиль языка
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
