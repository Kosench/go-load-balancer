package config

import (
	"github.com/spf13/viper"
	"strings"
)

type Config struct {
	ListenAddress string
	Backends      []string
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("ListenAddress", ":8080")
	viper.SetDefault("BACKENDS", []string{"localhost:9001", "localhost:9002"})

	viper.BindEnv("BACKENDS")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &Config{
		ListenAddress: viper.GetString("ListenAddress"),
		Backends:      strings.Split(viper.GetString("BACKENDS"), ","),
	}

	return cfg, nil
}
