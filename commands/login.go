package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func Login(cfg config.Config, api *services.API) *cli.Command {
	// nolint: forbidigo
	return &cli.Command{
		Name: "login",

		Action: func(ctx *cli.Context) error {
			dir, err := EnsureConfigDir()
			if err != nil {
				return fmt.Errorf("ensure config dir: %w", err)
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("username: ")

			username, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read string: %w", err)
			}

			fmt.Print("password: ")

			btsPassword, err := term.ReadPassword(syscall.Stdin)

			fmt.Println()

			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}

			cfg.Cookies, err = api.Login(ctx.Context, username, string(btsPassword))
			if err != nil {
				return fmt.Errorf("login: %w", err)
			}

			if err = config.Write(dir, config.Filename, cfg); err != nil {
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
