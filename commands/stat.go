package commands

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/kudrykv/go-vkpm/commands/before"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	flagYear = "year"
)

func Stat(cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name: "stat",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: flagYear},
		},
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			now := time.Now()
			year := c.Int(flagYear)
			month := time.December

			if now.Year() == year {
				month = now.Month()
			}

			group, cctx := errgroup.WithContext(c.Context)
			salariesChan := make(chan types.Salary, 2*int(month))
			historiesChan := make(chan types.ReportEntries, int(month))

			for i := time.January; i <= month; i++ {
				group.Go(getSalaryInChan(cctx, api, year, i, salariesChan))
				group.Go(getHistoryInChan(cctx, api, year, i, historiesChan))
			}

			if err := group.Wait(); err != nil {
				return fmt.Errorf("group: %w", err)
			}

			close(salariesChan)
			close(historiesChan)

			salaries := salariesChanToSlice(salariesChan)
			histories := historiesChanToSlice(historiesChan)

			_, _ = fmt.Fprintln(c.App.Writer, salaries)
			_, _ = fmt.Fprintln(c.App.Writer)
			_, _ = fmt.Fprintln(c.App.Writer, types.StatSalaryHistory{
				Year: year, Salaries: salaries, Histories: histories,
				StartMonth: time.January,
				EndMonth:   month,
			})

			return nil
		},
	}
}

func salariesChanToSlice(salariesChan chan types.Salary) types.Salaries {
	salaries := make(types.Salaries, 0, len(salariesChan))
	for salary := range salariesChan {
		salaries = append(salaries, salary)
	}

	sort.Slice(salaries, func(i, j int) bool {
		return salaries[i].Month < salaries[j].Month
	})

	return salaries
}

func getSalaryInChan(
	cctx context.Context, api *services.API, year int, m time.Month, salariesChan chan types.Salary,
) func() error {
	return func() error {
		salary, err := api.Salary(cctx, year, m)
		if err != nil {
			return fmt.Errorf("salary %d %v: %w", year, m, err)
		}

		salariesChan <- salary

		return nil
	}
}

func historiesChanToSlice(hc chan types.ReportEntries) types.GroupedEntries {
	out := make([]types.ReportEntries, 0, len(hc))

	for entries := range hc {
		out = append(out, entries)
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i][0].ReportDate.Month() < out[j][0].ReportDate.Month()
	})

	return out
}

func getHistoryInChan(
	cctx context.Context, api *services.API, year int, m time.Month, hc chan types.ReportEntries,
) func() error {
	return func() error {
		history, err := api.History(cctx, year, m)
		if err != nil {
			return fmt.Errorf("history %d %v: %w", year, m, err)
		}

		hc <- history

		return nil
	}
}
