package commands

import (
	"fmt"

	"github.com/kudrykv/go-vkpm/commands/before"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
)

func History(cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "history",
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			today := types.Today()
			history, err := api.History(c.Context, today.Year(), today.Month())
			if err != nil {
				return fmt.Errorf("history: %w", err)
			}

			_, _ = fmt.Fprintln(c.App.Writer, history)

			return nil
		},
	}
}
