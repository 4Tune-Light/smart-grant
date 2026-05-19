package config

import (
	"fmt"

	"github.com/spf13/viper"
)

const configPath = "configs"

func Load(cfgPath string) (*Config, error) {
	v := viper.New()

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if cfgPath != "" {
		v.AddConfigPath(cfgPath)
	} else {
		v.AddConfigPath(configPath)
	}

	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
