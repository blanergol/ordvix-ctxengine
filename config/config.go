package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	DefaultLevelDepth  = 100
	DefaultTimeReIndex = 2
)

type Context struct {
	ProjectDir      string
	ParseGenFile    bool
	SkipGenerated   bool
	ParseLevelDepth int
	TimeReIndex     int
	PrintIndex      bool
	WatchEnabled    bool
}

type Orchestrator struct {
	Host  string
	Token string
}

type Rag struct {
	Host  string
	Token string
}

type Config struct {
	Orchestrator Orchestrator
	Context      Context
	Rag          Rag
}

// NewConfig создаёт новую конфигурацию, используя viper для чтения из переменных окружения.
// Ключи соответствуют полям структур: orchestrator, context, rag.
func NewConfig() (*Config, error) {
	// Устанавливаем значения по умолчанию
	viper.SetDefault("orchestrator.host", "")
	viper.SetDefault("orchestrator.token", "")
	viper.SetDefault("context.project_dir", "")
	viper.SetDefault("context.parse_gen_file", false)
	viper.SetDefault("context.skip_generated", false)
	viper.SetDefault("context.parse_level_depth", DefaultLevelDepth)
	viper.SetDefault("context.time_reindex", DefaultTimeReIndex)
	viper.SetDefault("context.print_index", false)
	viper.SetDefault("context.watch_enabled", false)
	viper.SetDefault("rag.host", "")
	viper.SetDefault("rag.token", "")

	// Настраиваем viper для автоматического чтения переменных окружения.
	// Используем замену точек на нижнее подчеркивание.
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := &Config{
		Orchestrator: Orchestrator{
			Host:  viper.GetString("orchestrator.host"),
			Token: viper.GetString("orchestrator.token"),
		},
		Context: Context{
			ProjectDir:      viper.GetString("context.project_dir"),
			ParseGenFile:    viper.GetBool("context.parse_gen_file"),
			SkipGenerated:   viper.GetBool("context.skip_generated"),
			ParseLevelDepth: viper.GetInt("context.parse_level_depth"),
			TimeReIndex:     viper.GetInt("context.time_reindex"),
			PrintIndex:      viper.GetBool("context.print_index"),
			WatchEnabled:    viper.GetBool("context.watch_enabled"),
		},
		Rag: Rag{
			Host:  viper.GetString("rag.host"),
			Token: viper.GetString("rag.token"),
		},
	}

	return cfg, nil
}
