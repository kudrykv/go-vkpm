package commands

import (
	"errors"
	"fmt"
	"time"

	"github.com/kudrykv/go-vkpm/services"
	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
)

const (
	flagFor      = "for"
	flagProj     = "proj"
	flagSpan     = "span"
	flagFrom     = "from"
	flagTo       = "to"
	flagStatus   = "status"
	flagActivity = "activity"
)

var (
	errEmptyProj   = errors.New("empty project")
	errBadTime     = errors.New("must be rounded to 10 minutes")
	errRangeOrSpan = errors.New("use only range or span")
	errNoTime      = errors.New("specify span or time range")
)

func Report(cfg types.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name: "report",
		Flags: []cli.Flag{
			&cli.TimestampFlag{Name: flagFor, Layout: "01-02"},
			&cli.StringFlag{Name: flagProj},
			&cli.DurationFlag{Name: flagSpan},
			&cli.TimestampFlag{Layout: "15:04", Name: flagFrom},
			&cli.TimestampFlag{Layout: "15:04", Name: flagTo},
			&cli.IntFlag{Name: flagStatus, Value: 100},
			&cli.StringFlag{Name: flagActivity, Value: types.ActivityDevelopment},
		},
		Action: func(c *cli.Context) error {
			var (
				entry types.ReportEntry
				err   error
			)

			projects, err := api.Projects(c.Context)
			if err != nil {
				return fmt.Errorf("projects: %w", err)
			}

			if entry, err = parseEntry(c, cfg, projects); err != nil {
				return err
			}

			now := time.Now()
			history, err := api.History(c.Context, now.Year(), now.Month())
			if err != nil {
				return fmt.Errorf("history: %w", err)
			}

			if entry, err = entry.AlignTimes(history); err != nil {
				return fmt.Errorf("align: %w", err)
			}

			_, _ = fmt.Fprintln(c.App.Writer, entry)

			return nil
		},
	}
}

func parseEntry(c *cli.Context, cfg types.Config, projects types.Projects) (types.ReportEntry, error) {
	entry := types.ReportEntry{Project: types.Project{Name: c.String(flagProj)}}

	if len(entry.Project.Name) == 0 {
		if entry.Project.Name = cfg.DefaultProject; len(entry.Project.Name) == 0 {
			return entry, fmt.Errorf("no project provided: %w", errEmptyProj)
		}
	}

	var err error
	if entry, err = entry.UpdateProjectName(projects); err != nil {
		return entry, fmt.Errorf("fixup project name: %w", err)
	}

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
