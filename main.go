package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime/trace"
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
	ctx := context.Background()

	defer func() { os.Exit(code) }()

	err, stop := enabledTrace(ctx)
	if shouldExit(ctx, "enabled trace", err) {
		return
	}

	defer func() { shouldExit(ctx, "stop trace", stop()) }()

	ctx, task := trace.NewTask(ctx, "app")
	defer task.End()

	cfg, err := config.New(ctx, "", "")
	if shouldExit(ctx, "new config: %w", err) {
		return
	}

	var (
		p          = printer.Printer{W: os.Stdout, E: os.Stderr}
		httpClient = &http.Client{
			Transport: http.DefaultTransport,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Timeout: cfg.HTTPTimeout,
		}
	)

	api := services.NewAPI(httpClient, cfg).WithCookies(cfg.Cookies)

	app := &cli.App{
		Name:    "vkpm",
		Usage:   "cli tool to avoid clicking through VKPM UI",
		Version: "0.0.5",
		Commands: []*cli.Command{
			commands.Config(cfg),
			commands.Login(p, cfg, api),
			commands.Dashboard(p, cfg, api),
			commands.Report(p, cfg, api),
			commands.History(p, cfg, api),
			commands.Stat(p, cfg, api),
			commands.Vacations(p, cfg, api),
		},
	}

	shouldExit(ctx, "", app.RunContext(ctx, os.Args))
}

func enabledTrace(ctx context.Context) (error, func() error) {
	defer trace.StartRegion(ctx, "enable trace").End()

	noop := func() error { return nil }

	if os.Getenv("VKPM_ENABLE_TRACE") != "1" {
		return nil, noop
	}

	sock, err := os.Create(time.Now().Format("trace-20060102150405.out"))
	if err != nil {
		return fmt.Errorf("create trace file: %w", err), noop
	}

	if err := trace.Start(sock); err != nil {
		if err2 := sock.Close(); err2 != nil {
			return fmt.Errorf("trace start, close trace file: %v, %v", err, err2), noop
		}

		return fmt.Errorf("start trace: %w", err), noop
	}

	return nil, func() error {
		trace.Stop()

		if err := sock.Close(); err != nil {
			return fmt.Errorf("close trace file: %w", err)
		}

		return nil
	}
}

func shouldExit(ctx context.Context, msg string, err error) bool {
	defer trace.StartRegion(ctx, "should exit").End()

	if err == nil {
		return false
	}

	code = 1

	printErr(ctx, msg, err)

	return true
}

// nolint:forbidigo
func printErr(ctx context.Context, msg string, err error) {
	defer trace.StartRegion(ctx, "print err").End()

	if len(msg) > 0 {
		err = fmt.Errorf("%s: %w", msg, err)
	}

	indent := 0
	a, b := errors.Unwrap(err), err

	for a != nil {
		index := strings.Index(b.Error(), a.Error())

		fmt.Print(strings.Repeat(" ", indent))
		fmt.Println(b.Error()[0:index])
		indent += 2
		a, b = errors.Unwrap(a), a
	}

	fmt.Print(strings.Repeat(" ", indent))
	fmt.Println(b.Error())
}
