package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Domain         string  `yaml:"domain"`
	DefaultProject string  `yaml:"default_project"`
	Cookies        Cookies `yaml:"cookies"`
}

type Cookies struct {
	CSRFToken string `yaml:"csrftoken"`
	SessionID string `yaml:"sessionid"`
}

func EnsureDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}

	configRoot := filepath.Join(homeDir, ".config", "vkpm")

	if err = os.MkdirAll(configRoot, os.ModePerm); err != nil {
		return "", fmt.Errorf("mkdir all: %w", err)
	}

	return configRoot, nil
}

func Read(path, file string) (Config, error) {
	bts, err := ioutil.ReadFile(filepath.Join(path, file))
	if err != nil {
		return Config{}, fmt.Errorf("read file: %w", err)
	}

	var authConfig Config
	if err = yaml.Unmarshal(bts, &authConfig); err != nil {
		return authConfig, fmt.Errorf("unmarshal: %w", err)
	}

	return authConfig, nil
}

func Write(path, file string, config Config) error {
	bts, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = ioutil.WriteFile(filepath.Join(path, file), bts, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
