package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ListenAddress       string   `mapstructure:"ListenAddress"`
	Backends            []string `mapstructure:"Backends"`
	RateLimitCapacity   float64  `mapstructure:"RateLimitCapacity"`
	RateLimitRefillRate float64  `mapstructure:"RateLimitRefillRate"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("ListenAddress", ":8080")
	viper.SetDefault("BACKENDS", []string{"localhost:9001", "localhost:9002"})
	viper.SetDefault("RateLimitCapacity", 5.0)
	viper.SetDefault("RateLimitRefillRate", 1.0)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
