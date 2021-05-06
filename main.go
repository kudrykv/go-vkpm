package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kudrykv/go-vkpm/commands"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

func main() {
	ctx := context.Background()

	dir, err := EnsureConfigDir()
	if err != nil {
		exit("ensure config dir", err)
	}

	var config types.Config
	join := strings.Join([]string{dir, "config.yml"}, string(os.PathSeparator))

	if _, err = os.Stat(join); os.IsNotExist(err) {
		file, err := os.Create(join)
		if err != nil {
			exit("create", err)
		}

		if err = file.Close(); err != nil {
			exit("close", err)
		}
	} else {
		config, err = ReadConfig(dir, "config.yml")
		if err != nil {
			exit("read config", err)
		}
	}

	ctx = context.WithValue(ctx, types.Dir, dir)
	ctx = context.WithValue(ctx, types.Cfg, config)

	api := services.NewAPI(&http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}, config).
		WithCookies(config.Cookies)

	app := &cli.App{
		Name: "vkpm",
		Commands: []*cli.Command{
			commands.Config(),
			commands.Login(config, api),
			commands.Dashboard(api),
		},
	}

	if err = app.RunContext(ctx, os.Args); err != nil {
		exit("", err)
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

// nolint:forbidigo
func exit(msg string, err error) {
	if len(msg) == 0 {
		fmt.Println(err)
	} else {
		fmt.Println(msg+":", err)
	}

	os.Exit(1)
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
