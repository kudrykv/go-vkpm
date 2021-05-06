package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
)

func Login(config types.Config, api services.API) *cli.Command {
	return &cli.Command{
		Name: "login",

		Action: func(ctx *cli.Context) error {
			dir, err := EnsureConfigDir()
			if err != nil {
				return fmt.Errorf("ensure config dir: %w", err)
			}

			config.Cookies, err = api.Login(ctx.Context, "", "")
			if err != nil {
				return fmt.Errorf("login: %w", err)
			}

			if err = WriteConfig(dir, "config.yml", config); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

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
