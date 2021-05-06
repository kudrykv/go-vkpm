package commands

import (
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
			)

			thisMonth := time.Now()
			lastMonth := thisMonth.AddDate(0, -1, 0)

			group, cctx := errgroup.WithContext(c.Context)

			group.Go(func() error {
				var err error
				if thisMonthSalary, err = api.Salary(cctx, thisMonth.Year(), int(thisMonth.Month())); err != nil {
					return fmt.Errorf("this month salary: %w", err)
				}

				return nil
			})

			group.Go(func() error {
				var err error
				if lastMonthSalary, err = api.Salary(cctx, lastMonth.Year(), int(lastMonth.Month())); err != nil {
					return fmt.Errorf("last month salary: %w", err)
				}

				return nil
			})

			group.Go(func() error {
				var err error
				if historyEntries, err = api.History(cctx, thisMonth.Year(), thisMonth.Month()); err != nil {
					return fmt.Errorf("history: %w", err)
				}

				return nil
			})

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
