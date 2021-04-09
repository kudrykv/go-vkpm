package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func Config() *cli.Command {
	return &cli.Command{
		Name: "config",

		Before: func(ctx *cli.Context) error {
			path, err := EnsureConfigDir()
			if err != nil {
				return fmt.Errorf("ensure config dir: %w", err)
			}

			config, err := ReadConfig(path, "config.yml")
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			ctx.Context = context.WithValue(ctx.Context, types.Cfg, config)

			return nil
		},

		Action: func(ctx *cli.Context) error {
			value := ctx.Context.Value("test").(types.Config)

			fmt.Println(value)

			return nil
		},
	}
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

func ReadConfig(path, file string) (types.Config, error) {
	bts, err := ioutil.ReadFile(strings.Join([]string{path, file}, string(os.PathSeparator)))
	if err != nil {
		return types.Config{}, fmt.Errorf("read file: %w", err)
	}

	var authConfig types.Config
	if err = yaml.Unmarshal(bts, &authConfig); err != nil {
		return authConfig, fmt.Errorf("unmarshal: %w", err)
	}

	return authConfig, nil
}
