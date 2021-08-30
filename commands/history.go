package commands

import (
	"fmt"
	"time"

	"github.com/kudrykv/go-vkpm/commands/before"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/printer"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/th"
	"github.com/urfave/cli/v2"
)

func History(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:  "history",
		Usage: "show reported hours",
		Flags: []cli.Flag{
			&cli.TimestampFlag{
				Name: flagFor, Layout: "2006-01",
				DefaultText: "this month", Value: cli.NewTimestamp(time.Now()),
			},
		},
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "history")
			defer end()

			date := c.Timestamp(flagFor)

			history, err := api.History(ctx, date.Year(), date.Month())
			if err != nil {
				return fmt.Errorf("history in %d %v: %w", date.Year(), date.Month(), err)
			}

			p.Println(history)

			return nil
		},
	}
}
