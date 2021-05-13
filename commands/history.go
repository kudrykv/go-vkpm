package commands

import (
	"fmt"

	"github.com/kudrykv/vkpm/commands/before"
	"github.com/kudrykv/vkpm/config"
	"github.com/kudrykv/vkpm/printer"
	"github.com/kudrykv/vkpm/services"
	"github.com/kudrykv/vkpm/th"
	"github.com/kudrykv/vkpm/types"
	"github.com/urfave/cli/v2"
)

func History(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "history",
		Usage:  "show reported hours",
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "history")
			defer end()

			today := types.Today()
			history, err := api.History(ctx, today.Year(), today.Month())
			if err != nil {
				return fmt.Errorf("history in %d %v: %w", today.Year(), today.Month(), err)
			}

			p.Println(history)

			return nil
		},
	}
}
