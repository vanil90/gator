package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}

const configFileName = ".gatorconfig.json"

func Read() (Config, error) {
	var config Config
	configPath, err := getConfigFilePath()
	if err != nil {
		return config, err
	}

	configJson, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(configJson, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (c *Config) SetUser(username string) error {
	c.CurrentUsername = username
	err := write(*c)
	return err
}

func getConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	configPath := filepath.Join(home, configFileName)
	return configPath, nil
}

func write(cfg Config) error {
	configPath, err := getConfigFilePath()
	if err != nil {
		return err
	}

	file, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	err = enc.Encode(cfg)
	if err != nil {
		return err
	}

	return nil
}
