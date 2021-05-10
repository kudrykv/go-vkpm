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

	path string
	name string
}

type Cookies struct {
	CSRFToken string `yaml:"csrftoken"`
	SessionID string `yaml:"sessionid"`
}

func New(path, name string) (Config, error) {
	var (
		cfg  = Config{path: path, name: name}
		sock *os.File
		err  error
	)

	if len(cfg.path) == 0 {
		if cfg.path, err = EnsureDir(); err != nil {
			return cfg, fmt.Errorf("ensure dir: %w", err)
		}
	}

	if len(cfg.name) == 0 {
		cfg.name = Filename
	}

	join := filepath.Join(cfg.path, cfg.name)
	if _, err = os.Stat(join); os.IsNotExist(err) {
		sock, err = os.Create(join)
		if err != nil {
			return cfg, fmt.Errorf("create: %w", err)
		}

		if err = sock.Close(); err != nil {
			return cfg, fmt.Errorf("close: %w", err)
		}
	} else {
		cfg, err = cfg.Read()
		if err != nil {
			return cfg, fmt.Errorf("read: %w", err)
		}
	}

	return cfg, nil
}

func (c Config) Read() (Config, error) {
	bts, err := ioutil.ReadFile(filepath.Join(c.path, c.name))
	if err != nil {
		return c, fmt.Errorf("read file: %w", err)
	}

	if err = yaml.Unmarshal(bts, &c); err != nil {
		return c, fmt.Errorf("unmarshal: %w", err)
	}

	return c, nil
}

func (c Config) Write() error {
	bts, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = ioutil.WriteFile(filepath.Join(c.path, c.name), bts, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
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
