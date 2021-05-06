package commands

import (
	"github.com/kudrykv/go-vkpm/services"
	"github.com/urfave/cli/v2"
)

func Dashboard(api services.API) *cli.Command {
	return &cli.Command{
		Name: "dashboard",
		Action: func(c *cli.Context) error {
			api.Dashboard(c.Context)

			return nil
		},
	}
}
