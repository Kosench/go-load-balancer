package config

import (
	"errors"
	"fmt"

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

	// Try to read config file, but don't fail if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			// Return error only if it's not a "file not found" error
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found, using defaults
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.ListenAddress == "" {
		return errors.New("listen address cannot be empty")
	}

	if len(c.Backends) == 0 {
		return errors.New("at least one backend must be configured")
	}

	if c.RateLimitCapacity <= 0 {
		return errors.New("rate limit capacity must be greater than 0")
	}

	if c.RateLimitRefillRate <= 0 {
		return errors.New("rate limit refill rate must be greater than 0")
	}

	return nil
}
