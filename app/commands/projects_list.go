package commands

import (
	"fmt"

	"github.com/kudrykv/go-vkpm/app/commands/before"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/services"
	"github.com/urfave/cli/v2"
)

func ProjectsList(cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name: "list",

		Before: before.IsHTTPAuthMeet(cfg),

		Action: func(appCtx *cli.Context) error {
			projects, err := api.Projects(appCtx.Context)
			if err != nil {
				return fmt.Errorf("list projects: %w", err)
			}

			_, _ = fmt.Fprintln(appCtx.App.Writer, projects.String())

			return nil
		},
	}
}
