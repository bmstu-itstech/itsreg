package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

const projectPrefix = "IR"

type Config struct {
	HTTP        HTTP        `mapstructure:"http"`
	Logging     Logging     `mapstructure:"logging"`
	Proxy       Proxy       `mapstructure:"proxy"`
	Postgres    Postgres    `mapstructure:"postgres"`
	RateLimiter RateLimiter `mapstructure:"rate_limiter"`
	JWT         JWT         `mapstructure:"jwt"`
}

type HTTP struct {
	Port             int      `mapstructure:"port"`
	CORSAllowOrigins []string `mapstructure:"cors_allow_origins"`
	CORSMaxAge       int      `mapstructure:"cors_max_age"`
}

type Logging struct {
	Level string `mapstructure:"level"`
}

type Proxy struct {
	URL string `mapstructure:"url"`
}

type Postgres struct {
	URI string `mapstructure:"uri"`
}

type RateLimiter struct {
	Capacity int     `mapstructure:"capacity"`
	Rate     float64 `mapstructure:"rate"`
}

type JWT struct {
	Secret string `mapstructure:"secret"`
}

func Load(path string) (*Config, error) {
	// Нетривиальный момент Viper, не описанный в документации, но описанный в
	// 	https://github.com/spf13/viper/issues/1797
	// Без viper.ExperimentalBindStruct() переменная окружения загружается только если она была указана в yaml конфиге.
	// Так например:
	// 	postgres:
	//    uri:
	// и SC_POSTGRES_URI работает корректно, а без пустого uri - не читает переменную вовсе. Причём,
	//	viper.Get("postgres.uri")
	// работает исправно - проблема именно в Unmarshall, который почему-то полагается на файл.
	// Решение - viper.ExperimentalBindStruct().
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.SetEnvPrefix(projectPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config '%s': %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}
