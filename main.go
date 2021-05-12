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
	"github.com/kudrykv/go-vkpm/printer"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/urfave/cli/v2"
)

var code int

func main() {
	defer func() { os.Exit(code) }()

	ctx := context.Background()
	p := printer.Printer{W: os.Stdout, E: os.Stderr}

	cfg, err := config.New("", "")
	if shouldExit("new config: %w", err) {
		return
	}

	httpClient := &http.Client{
		Transport: http.DefaultTransport,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	api := services.NewAPI(httpClient, cfg).WithCookies(cfg.Cookies)

	app := &cli.App{
		Name: "vkpm",
		Commands: []*cli.Command{
			commands.Config(cfg),
			commands.Login(p, cfg, api),
			commands.Dashboard(p, cfg, api),
			commands.Report(p, cfg, api),
			commands.History(p, cfg, api),
			commands.Stat(p, cfg, api),
		},
	}

	shouldExit("", app.RunContext(ctx, os.Args))
}

// nolint:forbidigo
func shouldExit(msg string, err error) bool {
	if err == nil {
		return false
	}

	code = 1

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

	return true
}
