package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kudrykv/go-vkpm/commands"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()

	dir, err := EnsureConfigDir()
	if err != nil {
		exit("ensure config dir", err)
	}

	ctx = context.WithValue(ctx, types.Dir, dir)

	app := &cli.App{
		Name: "vkpm",
		Commands: []*cli.Command{
			commands.Config(),
		},
	}

	if err := app.RunContext(ctx, os.Args); err != nil {
		fmt.Println("error:", err) //nolint:forbidigo
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

func exit(msg string, err error) {
	fmt.Println(msg+":", err)
	os.Exit(1)
}
