package config

import (
	"github.com/spf13/viper"
)

type AppConfig struct {
	Client struct {
		PrivateKey string
	}
}

var (
	cfg *AppConfig
)

func Config() *AppConfig {
	if cfg == nil {
		load()
	}

	return cfg
}

func load() {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Ignore config file not found, perhaps we will use environment variables.
	_ = viper.ReadInConfig()

	cfg = &AppConfig{}

	// Read private key from config
	cfg.Client.PrivateKey = viper.GetString("CLIENT_PRIVATE_KEY")
}
