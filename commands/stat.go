package commands

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/kudrykv/vkpm/commands/before"
	"github.com/kudrykv/vkpm/config"
	"github.com/kudrykv/vkpm/printer"
	"github.com/kudrykv/vkpm/services"
	"github.com/kudrykv/vkpm/th"
	"github.com/kudrykv/vkpm/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func Stat(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:  "stat",
		Usage: "show money and hour stat for the given year",
		Flags: []cli.Flag{
			&cli.IntFlag{Name: flagFor, Usage: "year", Value: time.Now().Year()},
		},
		Before: before.IsHTTPAuthMeet(cfg),
		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "stat")
			defer end()

			startMonth := time.January
			endMonth := time.December
			year := c.Int(flagFor)
			now := time.Now()

			if now.Year() < year {
				return fmt.Errorf("future is unknown")
			}

			if now.Year() == year {
				endMonth = now.Month()
			}

			group, cctx := errgroup.WithContext(ctx)
			salariesChan := make(chan types.Salary, 2*int(endMonth))
			historiesChan := make(chan types.ReportEntries, int(endMonth))

			for i := startMonth; i <= endMonth; i++ {
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

			p.Println(types.StatSalaryHistory{
				Year: year, Salaries: salaries, Histories: histories,
				StartMonth: startMonth,
				EndMonth:   endMonth,
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
		if len(out[i]) == 0 {
			return true
		}

		if len(out[j]) == 0 {
			return false
		}

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
