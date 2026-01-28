package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	ApiKey         string `mapstructure:"API_KEY"`
	DatabaseURL    string `mapstructure:"DATABASE_URL"`
	AdminID        int    `mapstructure:"ADMIN_ID"`
	XrayConfigPath string `mapstructure:"XRAY_CONFIG_PATH"`
	ApiSslCertFile string `mapstructure:"API_SSL_CERTFILE"`
	ApiSslKeyFile  string `mapstructure:"API_SSL_KEYFILE"`
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
	if c.DatabaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	if c.AdminID == 0 {
		return errors.New("ADMIN_ID is required")
	}
	if c.XrayConfigPath == "" {
		return errors.New("XRAY_CONFIG_PATH is required")
	}
	if _, err := os.Stat(c.XrayConfigPath); os.IsNotExist(err) {
		return errors.New("XRAY_CONFIG_PATH file does not exist")
	}

	if (c.ApiSslCertFile != "" && c.ApiSslKeyFile == "") || (c.ApiSslCertFile == "" && c.ApiSslKeyFile != "") {
		return errors.New("both API_SSL_CERTFILE and API_SSL_KEYFILE must be set together")
	}
	if c.ApiSslCertFile != "" && c.ApiSslKeyFile != "" {
		if _, err := os.Stat(c.ApiSslCertFile); os.IsNotExist(err) {
			return errors.New("API_SSL_CERTFILE does not exist:" + c.ApiSslCertFile)
		}
		if _, err := os.Stat(c.ApiSslKeyFile); os.IsNotExist(err) {
			return errors.New("API_SSL_KEYFILE does not exist:" + c.ApiSslKeyFile)
		}
	}
	return nil
}
