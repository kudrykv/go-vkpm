package commands

import (
	"fmt"
	"time"

	"github.com/kudrykv/vkpm/commands/before"
	"github.com/kudrykv/vkpm/config"
	"github.com/kudrykv/vkpm/printer"
	"github.com/kudrykv/vkpm/services"
	"github.com/kudrykv/vkpm/th"
	"github.com/urfave/cli/v2"
)

func Vacations(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "vacations",
		Usage:  "show requested vacations",
		Before: before.IsHTTPAuthMeet(cfg),
		Flags: []cli.Flag{
			&cli.IntFlag{Name: flagYear, Value: time.Now().Year()},
		},
		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "vacations")
			defer end()

			paidDays, vacations, _, err := api.VacationsHolidays(ctx, c.Int(flagYear))
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
