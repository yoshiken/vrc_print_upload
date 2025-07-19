package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	APIBaseURL string
	configDir  string
}

func Load(cfgFile string) (*Config, error) {
	cfg := &Config{
		APIBaseURL: "https://api.vrchat.cloud/api/1",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cfg.configDir = filepath.Join(homeDir, ".vrc-print")
	if err := os.MkdirAll(cfg.configDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(cfg.configDir)
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("VRC_PRINT")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	}

	if apiURL := viper.GetString("api_base_url"); apiURL != "" {
		cfg.APIBaseURL = apiURL
	}

	return cfg, nil
}

func (c *Config) ConfigDir() string {
	return c.configDir
}

func (c *Config) CookieFile() string {
	return filepath.Join(c.configDir, "cookies.json")
}