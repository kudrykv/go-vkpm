package commands

import (
	"fmt"
	"runtime/trace"

	"github.com/kudrykv/go-vkpm/commands/before"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/printer"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
)

func History(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "history",
		Usage:  "show reported hours",
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			ctx, task := trace.NewTask(c.Context, "history")
			defer task.End()

			today := types.Today()
			history, err := api.History(ctx, today.Year(), today.Month())
			if err != nil {
				return fmt.Errorf("history: %w", err)
			}

			p.Println(history)

			return nil
		},
	}
}
