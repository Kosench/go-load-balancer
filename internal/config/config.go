package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ListenAddress       string   `mapstructure:"LISTEN_ADDRESS"`
	Backends            []string `mapstructure:"BACKENDS"`
	RateLimitCapacity   float64  `mapstructure:"RATE_LIMIT_CAPACITY"`
	RateLimitRefillRate float64  `mapstructure:"RATE_LIMIT_REFILL_RATE"`
}

func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	viper.SetDefault("LISTEN_ADDRESS", ":8080")
	viper.SetDefault("BACKENDS", []string{"localhost:9001", "localhost:9002"})
	viper.SetDefault("RATE_LIMIT_CAPACITY", 5.0)
	viper.SetDefault("RATE_LIMIT_REFILL_RATE", 1.0)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
