package types

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

const (
	ActivityEstimate    = "estimate"
	ActivityDevelopment = "development"
	ActivityTesting     = "testing"
	ActivityBugfixing   = "bugfixing"
	ActivityManagement  = "management"
	ActivityAnalysis    = "analysis"
)

type ReportEntries []ReportEntry

func (e ReportEntries) Reported(d Date) bool {
	for _, ee := range e {
		if ee.ReportDate.Year() == d.Year() && ee.ReportDate.Month() == d.Month() && ee.ReportDate.Day() == d.Day() {
			return true
		}
	}

	return false
}

type ReportEntry struct {
	ID          string
	PublishDate Date
	ReportDate  Date
	ProjectName string
	Activity    string
	Name        string
	Description string
	Status      int
	StartTime   time.Time
	EndTime     time.Time
	Span        time.Duration
}

func (e ReportEntry) IsEmptyRange() bool {
	return e.StartTime.IsZero() && e.EndTime.IsZero()
}

func (e ReportEntry) IsSpanAndRangePresent() bool {
	return !e.StartTime.IsZero() && !e.EndTime.IsZero() && e.Span > 0
}

func (e ReportEntry) IsSpanAndRangeAbsent() bool {
	return e.StartTime.IsZero() && e.EndTime.IsZero() && e.Span == 0
}

var (
	ErrNoStartTime    = errors.New("no start time")
	ErrNoEndTime      = errors.New("no end time")
	ErrStartLargerEnd = errors.New("start time is greater or equal than end time")
	ErrTimeNotRounded = errors.New("time must be rounded to 10 minutes")
	ErrStatusNegative = errors.New("status is less than zero")
	ErrStatusMore100  = errors.New("status is larger than 100")
	ErrStatusNotRound = errors.New("must be rounded to 10")
	ErrBadActivity    = errors.New("bad activity")
)

func (e ReportEntry) TestBrokenRange() error {
	if e.IsEmptyRange() {
		return nil
	}

	if e.StartTime.IsZero() && !e.EndTime.IsZero() {
		return ErrNoStartTime
	}

	if !e.StartTime.IsZero() && e.EndTime.IsZero() {
		return ErrNoEndTime
	}

	if e.StartTime.After(e.EndTime) || e.StartTime.Equal(e.EndTime) {
		return ErrStartLargerEnd
	}

	if !e.StartTime.Round(10 * time.Minute).Equal(e.StartTime) {
		return fmt.Errorf("start time: %w", ErrTimeNotRounded)
	}

	if !e.EndTime.Round(10 * time.Minute).Equal(e.EndTime) {
		return fmt.Errorf("end time: %w", ErrTimeNotRounded)
	}

	return nil
}

func (e ReportEntry) TestStatus() error {
	if e.Status < 0 {
		return ErrStatusNegative
	}

	if e.Status > 100 {
		return ErrStatusMore100
	}

	if e.Status%10 != 0 {
		return ErrStatusNotRound
	}

	return nil
}

func (e ReportEntry) SetActivity(short string) (ReportEntry, error) {
	activities := []string{
		ActivityEstimate, ActivityDevelopment, ActivityTesting, ActivityBugfixing, ActivityManagement, ActivityAnalysis,
	}

	for _, activity := range activities {
		if idx := strings.Index(strings.ToLower(activity), strings.ToLower(short)); idx != 0 {
			continue
		}

		e.Activity = activity

		return e, nil
	}

	return e, fmt.Errorf("%v: %w", short, ErrBadActivity)
}

func (e ReportEntry) UpdateProjectName(available Projects) (ReportEntry, error) {
	project, err := available.Match(e.ProjectName)
	if err != nil {
		return e, fmt.Errorf("match: %w", err)
	}

	e.ProjectName = project.Name

	return e, nil
}

func NewReportEntriesFromHTMLNode(doc *html.Node) (ReportEntries, error) {
	nodes, err := htmlquery.QueryAll(doc, `//table[@id="history"]//tbody/tr`)
	if err != nil {
		return nil, fmt.Errorf("query all: %w", err)
	}

	entries := make(ReportEntries, 0, len(nodes))

	for _, node := range nodes {
		var entry ReportEntry

		iter := []struct {
			s    *string
			expr string
		}{
			{&entry.ID, `./td[1]`},
			{&entry.ProjectName, `./td[4]`},
			{&entry.Activity, `./td[5]//option[@selected]`},
			{&entry.Name, `./td[6]`},
			{&entry.Description, `./td[7]`},
		}

		for _, kv := range iter {
			if *kv.s, err = getTextFromNode(node, kv.expr); err != nil {
				return nil, fmt.Errorf("get text from node: %w", err)
			}
		}

		iter2 := []struct {
			s      *time.Time
			layout string
			expr   string
		}{

			{&entry.StartTime, `15:04`, `./td[9]`},
			{&entry.EndTime, `15:04`, `./td[10]`},
		}

		for _, kv := range iter2 {
			if *kv.s, err = getTimeFromNode(node, kv.layout, kv.expr); err != nil {
				return nil, fmt.Errorf("get time from node: %w", err)
			}
		}

		iter3 := []struct {
			s      *Date
			layout string
			expr   string
		}{
			{&entry.PublishDate, `2 Jan, Mon 15:04`, `./td[2]`},
			{&entry.ReportDate, `2 Jan, 2006`, `./td[3]`},
		}

		for _, kv := range iter3 {
			if *kv.s, err = getDateFromNode(node, kv.layout, kv.expr); err != nil {
				return nil, fmt.Errorf("get time from node: %w", err)
			}
		}

		if entry.Status, err = getIntFromNode(node, `./td[8]`); err != nil {
			return nil, fmt.Errorf("get int from node: %w", err)
		}

		text, err := getTextFromNode(node, `./td[11]`)
		if err != nil {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		if entry.Span, err = time.ParseDuration(text); err != nil {
			return nil, fmt.Errorf("parse duration: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
