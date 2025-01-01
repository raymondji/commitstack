package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Theme ThemeConfig `json:"theme"`

	// Keys are the git repo path, e.g. "raymondji/commitstack"
	Repositories map[string]RepoConfig `json:"repositories"`
}

type ThemeConfig struct {
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	TertiaryColor  string `json:"tertiaryColor"`
}

type RepoConfig struct {
	DefaultBranch string       `json:"defaultBranch"`
	Gitlab        GitlabConfig `json:"gitlab"`
	Github        GithubConfig `json:"github"`
}

type GitlabConfig struct {
	PersonalAccessToken string `json:"personalAccessToken"`
}

type GithubConfig struct {
	PersonalAccessToken string `json:"personalAccessToken"`
}

func Load() (*Config, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}

	viper.SetConfigFile(configFilePath)
	viper.SetConfigType("json")

	if err := viper.ReadInConfig(); err != nil {
		var viperNotFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &viperNotFoundErr) || errors.Is(err, os.ErrNotExist) {
			// Return defaults
			return &Config{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// Save returns (config file path, error)
func Save(cfg *Config) (string, error) {
	configFilePath, err := getConfigFilePath()
	if err != nil {
		return "", err
	}

	// Marshal the config struct into JSON format.
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write the JSON data to the configuration file.
	if err := os.WriteFile(configFilePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write config file: %w", err)
	}

	return configFilePath, nil
}

// getConfigFilePath constructs the config file path in the user's home directory.
func getConfigFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".commitstack.json"), nil
}
