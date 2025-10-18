package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DBPath string `yaml:"db_path"`
}

var globalConfig *Config

const (
	defaultConfigDir  = "~/.mngr"
	defaultConfigFile = "~/.mngr/config.yaml"
	defaultDBPath     = "~/.mngr/main.db"
)

func Init() error {
	configPath := expandPath(defaultConfigFile)
	configDir := expandPath(defaultConfigDir)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		if err := createDefaultConfig(configPath); err != nil {
			return err
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	globalConfig = &cfg
	return nil
}

func createDefaultConfig(path string) error {
	defaultCfg := Config{
		DBPath: defaultDBPath,
	}

	data, err := yaml.Marshal(defaultCfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func GetDBPath() string {
	if globalConfig == nil {
		return defaultDBPath
	}
	return expandPath(globalConfig.DBPath)
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[1:])
	}
	return path
}
