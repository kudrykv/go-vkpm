package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kudrykv/go-vkpm/commands/before"
	"github.com/kudrykv/go-vkpm/config"
	"github.com/kudrykv/go-vkpm/printer"
	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	flagFor      = "for"
	flagProj     = "proj"
	flagSpan     = "span"
	flagFrom     = "from"
	flagTo       = "to"
	flagStatus   = "status"
	flagActivity = "activity"
	flagTitle    = "title"
	flagMessage  = "message"
)

var (
	errEmptyProj   = errors.New("empty project")
	errBadTime     = errors.New("must be rounded to 10 minutes")
	errRangeOrSpan = errors.New("use only range or span")
	errNoTime      = errors.New("specify span or time range")
)

func Report(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:   "report",
		Before: before.IsHTTPAuthMeet(cfg),
		Flags: []cli.Flag{
			&cli.TimestampFlag{Name: flagFor, Layout: "01-02", DefaultText: "not set"},
			&cli.StringFlag{Name: flagProj},
			&cli.DurationFlag{Name: flagSpan, DefaultText: "not set"},
			&cli.TimestampFlag{Layout: "15:04", Name: flagFrom, DefaultText: "not set"},
			&cli.TimestampFlag{Layout: "15:04", Name: flagTo, DefaultText: "not set"},
			&cli.IntFlag{Name: flagStatus, Value: 100},
			&cli.StringFlag{Name: flagActivity, Value: types.ActivityDevelopment},
			&cli.StringFlag{Name: flagTitle},
			&cli.StringFlag{Name: flagMessage, Aliases: []string{"m"}, Required: true},
		},
		Action: func(c *cli.Context) error {
			var (
				history  types.ReportEntries
				projects types.Projects
				entry    types.ReportEntry
				err      error
			)

			if entry, err = parseEntry(c, cfg); err != nil {
				return err
			}

			today := types.Today()
			group, cctx := errgroup.WithContext(c.Context)

			group.Go(getHistory(cctx, api, today, &history))
			group.Go(getProjects(cctx, api, &projects))

			if err = group.Wait(); err != nil {
				return fmt.Errorf("group: %w", err)
			}

			if entry, err = entry.UpdateProjectName(projects); err != nil {
				return fmt.Errorf("fixup project name: %w", err)
			}

			if entry, err = entry.AlignTimes(history); err != nil {
				return fmt.Errorf("align: %w", err)
			}

			if entry, err = api.Report(c.Context, entry); err != nil {
				return fmt.Errorf("report: %w", err)
			}

			p.Println(entry)

			return nil
		},
	}
}

func getProjects(cctx context.Context, api *services.API, projects *types.Projects) func() error {
	return func() error {
		var err error
		if *projects, err = api.Projects(cctx); err != nil {
			return fmt.Errorf("projects: %w", err)
		}

		return nil
	}
}

func parseEntry(c *cli.Context, cfg config.Config) (types.ReportEntry, error) {
	entry := types.ReportEntry{
		Project:     types.Project{Name: c.String(flagProj)},
		Name:        c.String(flagTitle),
		Description: c.String(flagMessage),
	}

	if len(entry.Project.Name) == 0 {
		if entry.Project.Name = cfg.DefaultProject; len(entry.Project.Name) == 0 {
			return entry, fmt.Errorf("no project provided: %w", errEmptyProj)
		}
	}

	var err error
	if entry, err = entry.SetActivity(c.String(flagActivity)); err != nil {
		return entry, fmt.Errorf("test activity: %w", err)
	}

	entry.Status = c.Int(flagStatus)
	if err := entry.TestStatus(); err != nil {
		return entry, fmt.Errorf("test status: %w", err)
	}

	entry.ReportDate = types.Today()
	if tmp := c.Timestamp(flagFor); tmp != nil && !tmp.IsZero() {
		entry.ReportDate = types.Date{Time: *tmp}
	}

	if entry.Span = c.Duration(flagSpan); entry.Span > 0 {
		if entry.Span != entry.Span.Round(10*time.Minute) {
			return entry, fmt.Errorf("%v: %w", entry.Span, errBadTime)
		}
	}

	if tmp := c.Timestamp(flagFrom); tmp != nil && !tmp.IsZero() {
		entry.StartTime = *tmp
	}

	if tmp := c.Timestamp(flagTo); tmp != nil && !tmp.IsZero() {
		entry.EndTime = *tmp
	}

	if err := entry.TestBrokenRange(); err != nil {
		return entry, fmt.Errorf("broken range: %w", err)
	}

	if entry.IsSpanAndRangePresent() {
		return entry, fmt.Errorf("both range and from-to defined: %w", errRangeOrSpan)
	}

	if entry.IsSpanAndRangeAbsent() {
		return entry, fmt.Errorf("no time: %w", errNoTime)
	}

	return entry, nil
}
