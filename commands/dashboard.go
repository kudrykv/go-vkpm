package commands

import (
	"fmt"
	"time"

	"github.com/kudrykv/go-vkpm/services"
	"github.com/urfave/cli/v2"
)

func Dashboard(api *services.API) *cli.Command {
	return &cli.Command{
		Name: "dashboard",
		Action: func(c *cli.Context) error {
			now := time.Now()

			salary, err := api.Salary(c.Context, now.Year(), int(now.Month()))
			if err != nil {
				return fmt.Errorf("salary: %w", err)
			}

			_ = salary

			if err = api.Birthdays(c.Context); err != nil {
				return fmt.Errorf("birthdays: %w", err)
			}

			return nil
		},
	}
}
