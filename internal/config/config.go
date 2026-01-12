package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	configDirName  = "llmv"
	configFileName = "config.yaml"
)

type Config struct {
	Host    string `yaml:"host"`
	Storage struct {
		Path string `yaml:"path"`
	} `yaml:"storage"`
	Assistant struct {
		Message string `yaml:"message"`
	} `yaml:"system"`
}

func LoadOrNew(cliHost string) (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		cfg, err := createDefaultConfig(configPath)
		if err != nil {
			return nil, err
		}
		if cliHost != "" {
			cfg.Host = cliHost
		}
		return cfg, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if cliHost != "" {
		cfg.Host = cliHost
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) SetSystemMessage(msg string) error {
	c.Assistant.Message = msg
	return Save(c)
}

func createDefaultConfig(path string) (*Config, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}

	historyPath := filepath.Join(userHomeDir, "."+configDirName, "history")

	cfg := &Config{}
	cfg.Storage.Path = historyPath
	cfg.Host = "http://localhost:11434"
	cfg.Assistant.Message = ""

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal default config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write default config file: %w", err)
	}

	return cfg, nil
}

func getConfigPath() (string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home dir: %w", err)
	}
	return filepath.Join(userHomeDir, "."+configDirName, configFileName), nil
}
