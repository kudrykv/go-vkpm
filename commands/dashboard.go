package commands

import (
	"context"
	"fmt"
	"runtime/trace"

	"github.com/kudrykv/go-vkpm/commands/before"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/printer"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func Dashboard(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "dashboard",
		Usage:  "see stats for the current month",
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			ctx, task := trace.NewTask(c.Context, "dashboard")
			defer task.End()

			var (
				thisMonthSalary types.Salary
				lastMonthSalary types.Salary
				history         types.ReportEntries
				holidays        types.Holidays
				vacations       types.Vacations

				thisMonth = types.Today()
				lastMonth = thisMonth.AddDate(0, -1, 0)
			)

			group, cctx := errgroup.WithContext(ctx)

			group.Go(getSalary(cctx, api, thisMonth, &thisMonthSalary))
			group.Go(getSalary(cctx, api, lastMonth, &lastMonthSalary))
			group.Go(getHistory(cctx, api, thisMonth, &history))
			group.Go(getVacationsHolidays(cctx, api, lastMonth, &vacations, &holidays))

			if err := group.Wait(); err != nil {
				return fmt.Errorf("group: %w", err)
			}

			p.Println(thisMonthSalary.StringTotalPaid())
			p.Println(lastMonthSalary.StringTotalPaid())
			p.Println()
			p.Println(types.NewMonthInfo(thisMonth, thisMonthSalary, vacations, holidays, history))

			return nil
		},
	}
}

func getVacationsHolidays(
	cctx context.Context, api *services.API, moment types.Date, vacations *types.Vacations, holidays *types.Holidays,
) func() error {
	return func() error {
		defer trace.StartRegion(cctx, "get vacation holidays").End()

		var err error
		if *vacations, *holidays, err = api.VacationsHolidays(cctx, moment.Year()); err != nil {
			return fmt.Errorf("vacations: %w", err)
		}

		return nil
	}
}

func getHistory(cctx context.Context, api *services.API, moment types.Date, entries *types.ReportEntries) func() error {
	return func() error {
		defer trace.StartRegion(cctx, "get history").End()

		var err error
		if *entries, err = api.History(cctx, moment.Year(), moment.Month()); err != nil {
			return fmt.Errorf("history in %d %v: %w", moment.Year(), moment.Month(), err)
		}

		return nil
	}
}

func getSalary(cctx context.Context, api *services.API, moment types.Date, salary *types.Salary) func() error {
	return func() error {
		defer trace.StartRegion(cctx, "get salary").End()

		var err error
		if *salary, err = api.Salary(cctx, moment.Year(), moment.Month()); err != nil {
			return fmt.Errorf("salary in %d %v: %w", moment.Year(), moment.Month(), err)
		}

		return nil
	}
}
