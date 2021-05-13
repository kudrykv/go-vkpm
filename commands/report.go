package commands

import (
	"context"
	"errors"
	"fmt"
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
		Name:  "report",
		Usage: "report time",
		Description: "" +
			"Report time spent.\n\n" +
			"One need to specify the project, time frame and a message as a minimal number of parameters:\n\n" +
			"    vkpm report --proj projname --span 2h -m 'developing custom cli tools'\n\n" +
			"Time can be specifed as a span (--span 2h), or a from-to range (--from 10:00 --to 12:00).\n" +
			"If specified as a span, reports start from 09:00 of the given day, and stack one at each other.\n\n" +
			"    vkpm report --proj projname --span 2h -m 'dev cli tools'         # 09:00-11:00\n" +
			"    vkpm report --proj projname --span 2h -m 'doing other dev stuff' # 11:00-13:00\n\n",
		Before: before.IsHTTPAuthMeet(cfg),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: flagProj, Aliases: []string{"p"},
				Usage: "report for the specified project. Use default if not set",
			},
			&cli.TimestampFlag{
				Name: flagFor, Layout: "01-02", DefaultText: "today", Aliases: []string{"F"},
				Usage: "report date, specified in format MM-DD",
			},
			&cli.DurationFlag{
				Name: flagSpan, DefaultText: "not set", Aliases: []string{"s"},
				Usage: "time spent, e.g., 1h or 2h30m",
			},
			&cli.TimestampFlag{
				Layout: "15:04", Name: flagFrom, DefaultText: "not set", Aliases: []string{"f"},
				Usage: "start of the time range in format HH:MM",
			},
			&cli.TimestampFlag{
				Layout: "15:04", Name: flagTo, DefaultText: "not set", Aliases: []string{"t"},
				Usage: "end of the time range",
			},
			&cli.IntFlag{
				Name: flagStatus, Value: 100, Aliases: []string{"S"},
				Usage: "completeness, multiple of ten -- 0, 10, ..., 100",
			},
			&cli.StringFlag{
				Name: flagActivity, Value: types.ActivityDevelopment, Aliases: []string{"a"},
				Usage: "estimate, development, testing, bugfixing, management, analysis",
			},
			&cli.StringFlag{Name: flagTitle, Aliases: []string{"T"}, Usage: "report title", DefaultText: "project name"},
			&cli.StringFlag{
				Name: flagMessage, Aliases: []string{"m"}, Required: true,
				Usage: "what did you do in the given time frame",
			},
		},
		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "report")
			defer end()

			var (
				history  types.ReportEntries
				projects types.Projects
				entry    types.ReportEntry
				err      error
			)

			if entry, err = parseEntry(c, cfg); err != nil {
				return fmt.Errorf("parse entry: %w", err)
			}

			today := types.Today()
			group, cctx := errgroup.WithContext(ctx)

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
		return entry, fmt.Errorf("set activity %s: %w", c.String(flagActivity), err)
	}

	entry.Status = c.Int(flagStatus)
	if err := entry.TestStatus(); err != nil {
		return entry, fmt.Errorf("test status %d: %w", c.Int(flagStatus), err)
	}

	entry.ReportDate = types.Today()
	if tmp := c.Timestamp(flagFor); tmp != nil && !tmp.IsZero() {
		entry.ReportDate = types.Date{Time: *tmp}
	}

	if entry.Span = c.Duration(flagSpan); entry.Span > 0 {
		if entry.Span != entry.Span.Round(10*time.Minute) {
			return entry, fmt.Errorf("span %v: %w", entry.Span, errBadTime)
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
		return entry, fmt.Errorf("both range and span defined: %w", errRangeOrSpan)
	}

	if entry.IsSpanAndRangeAbsent() {
		return entry, fmt.Errorf("no range nor span defined: %w", errNoTime)
	}

	return entry, nil
}
