package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kudrykv/go-vkpm/commands"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()

	cfg, err := config.New("", "")
	if err != nil {
		exit("new config: %w", err)
	}

	api := services.NewAPI(&http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 5 * time.Second,
	}, cfg).
		WithCookies(cfg.Cookies)

	app := &cli.App{
		Name: "vkpm",
		Commands: []*cli.Command{
			commands.Config(cfg),
			commands.Login(cfg, api),
			commands.Dashboard(cfg, api),
			commands.Report(cfg, api),
		},
	}

	if err = app.RunContext(ctx, os.Args); err != nil {
		exit("", err)
	}
}

// nolint:forbidigo
func exit(msg string, err error) {
	if len(msg) > 0 {
		err = fmt.Errorf("%s: %w", msg, err)
	}

	split := strings.Split(err.Error(), ":")

	indent := 0
	for i, msg := range split {
		fmt.Print(strings.Repeat(" ", indent))

		if i+1 == len(split) {
			fmt.Println(msg)
		} else {
			fmt.Println(msg + ":")
		}

		indent += 2
	}

	os.Exit(1)
}
