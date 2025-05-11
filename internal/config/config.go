package config

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	ListenAddress string   `mapstructure:"ListenAddress"`
	Backends      []string `mapstructure:"UPSTREAMS"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	var cfg Config

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	cfg.Backends = strings.Split(viper.GetString("BACKENDS"), ",")
	return &cfg, nil
}
