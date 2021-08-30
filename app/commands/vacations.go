package commands

import (
	"fmt"
	"time"

	"github.com/kudrykv/go-vkpm/app/commands/before"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/printer"
	"github.com/kudrykv/go-vkpm/app/services"
	"github.com/kudrykv/go-vkpm/app/th"
	"github.com/urfave/cli/v2"
)

func Vacations(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "vacations",
		Usage:  "show requested and approved vacations",
		Before: before.IsHTTPAuthMeet(cfg),
		Flags: []cli.Flag{
			&cli.IntFlag{Name: flagFor, Usage: "year", Value: time.Now().Year()},
		},
		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "vacations")
			defer end()

			paidDays, vacations, _, err := api.VacationsHolidays(ctx, c.Int(flagFor))
			if err != nil {
				return fmt.Errorf("vacations holidays: %w", err)
			}

			_ = paidDays
			_ = vacations

			p.Printf("%v day(s) of paid vacations left\n\n", paidDays)
			p.Println(vacations)

			return nil
		},
	}
}
