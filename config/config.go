package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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

func EnsureConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user home dir: %w", err)
	}

	configRoot := strings.Join([]string{homeDir, ".config", "vkpm"}, string(os.PathSeparator))

	if err = os.MkdirAll(configRoot, os.ModePerm); err != nil {
		return "", fmt.Errorf("mkdir all: %w", err)
	}

	return configRoot, nil
}

func ReadConfig(path, file string) (Config, error) {
	bts, err := ioutil.ReadFile(strings.Join([]string{path, file}, string(os.PathSeparator)))
	if err != nil {
		return Config{}, fmt.Errorf("read file: %w", err)
	}

	var authConfig Config
	if err = yaml.Unmarshal(bts, &authConfig); err != nil {
		return authConfig, fmt.Errorf("unmarshal: %w", err)
	}

	return authConfig, nil
}

func WriteConfig(path, file string, config Config) error {
	bts, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = ioutil.WriteFile(strings.Join([]string{path, file}, string(os.PathSeparator)), bts, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
