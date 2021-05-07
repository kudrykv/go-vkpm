package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func Dashboard(api *services.API) *cli.Command {
	return &cli.Command{
		Name: "dashboard",
		Action: func(c *cli.Context) error {
			var (
				thisMonthSalary types.Salary
				lastMonthSalary types.Salary
				historyEntries  types.ReportEntries
				holidays        types.Holidays
				vacations       types.Vacations
			)

			thisMonth := time.Now()
			lastMonth := thisMonth.AddDate(0, -1, 0)

			group, cctx := errgroup.WithContext(c.Context)

			group.Go(getSalary(cctx, api, thisMonth, &thisMonthSalary))
			group.Go(getSalary(cctx, api, lastMonth, &lastMonthSalary))
			group.Go(getHistory(cctx, api, thisMonth, &historyEntries))
			group.Go(getVacationsHolidays(cctx, api, lastMonth, &vacations, &holidays))

			if err := group.Wait(); err != nil {
				return fmt.Errorf("group: %w", err)
			}

			_, _ = fmt.Fprintln(c.App.Writer, thisMonthSalary.StringTotalPaid())
			_, _ = fmt.Fprintln(c.App.Writer, lastMonthSalary.StringTotalPaid())
			_, _ = fmt.Fprintln(c.App.Writer)
			_, _ = fmt.Fprintln(c.App.Writer, thisMonthSalary.StringHoursReport())

			_, _ = fmt.Fprintln(c.App.Writer, historyEntries)

			return nil
		},
	}
}

func getVacationsHolidays(
	cctx context.Context, api *services.API, moment time.Time, vacations *types.Vacations, holidays *types.Holidays,
) func() error {
	return func() error {
		var err error
		if *vacations, *holidays, err = api.VacationsHolidays(cctx, moment.Year()); err != nil {
			return fmt.Errorf("vacations: %w", err)
		}

		return nil
	}
}

func getHistory(cctx context.Context, api *services.API, moment time.Time, entries *types.ReportEntries) func() error {
	return func() error {
		var err error
		if *entries, err = api.History(cctx, moment.Year(), moment.Month()); err != nil {
			return fmt.Errorf("history: %w", err)
		}

		return nil
	}
}

func getSalary(cctx context.Context, api *services.API, moment time.Time, salary *types.Salary) func() error {
	return func() error {
		var err error
		if *salary, err = api.Salary(cctx, moment.Year(), int(moment.Month())); err != nil {
			return fmt.Errorf("salary: %w", err)
		}

		return nil
	}
}
