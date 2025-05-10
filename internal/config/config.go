package config

import "github.com/spf13/viper"

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

	err = viper.Unmarshal(&cfg)
	return &cfg, nil
}
