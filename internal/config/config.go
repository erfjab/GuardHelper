package config

import (
	"errors"
	"github.com/spf13/viper"
)

type Config struct {
	ApiKey string `mapstructure:"API_KEY"`
}

var Cfg *Config

func LoadConfig() (*Config, error) {
	var config Config

	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return &config, err
	}

	err = config.Validate()
	if err != nil {
		return &config, err
	}

	Cfg = &config
	return &config, nil

}

func (c *Config) Validate() error {
	if c.ApiKey == "" {
		return errors.New("API_KEY is required")
	}
	return nil
}