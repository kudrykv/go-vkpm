package types

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/vkpm/th"
	"golang.org/x/net/html"
)

const (
	ActivityEstimate    = "estimate"
	ActivityDevelopment = "development"
	ActivityTesting     = "testing"
	ActivityBugfixing   = "bugfixing"
	ActivityManagement  = "management"
	ActivityAnalysis    = "analysis"

	ActivityENumEstimate    = "0"
	ActivityENumDevelopment = "1"
	ActivityENumTesting     = "2"
	ActivityENumBugfixing   = "3"
	ActivityENumManagement  = "4"
	ActivityENumAnalysis    = "5"
)

var (
	mapActivityEnum = map[string]string{
		ActivityEstimate:    ActivityENumEstimate,
		ActivityDevelopment: ActivityENumDevelopment,
		ActivityTesting:     ActivityENumTesting,
		ActivityBugfixing:   ActivityENumBugfixing,
		ActivityManagement:  ActivityENumManagement,
		ActivityAnalysis:    ActivityENumAnalysis,
	}
)

type GroupedEntries []ReportEntries

func (g GroupedEntries) StringYearView() string {
	s := make([]string, 0, len(g))

	for _, monthly := range g {
		month := monthly[0].ReportDate.Format("January")
		s = append(s, month+": "+monthly.ProjectHours().String())
	}

	return strings.Join(s, "\n")
}

func (g GroupedEntries) At(year int, month time.Month) ReportEntries {

	for _, entries := range g {
		if len(entries) == 0 {
			continue
		}

		if entries[0].ReportDate.Year() == year && entries[0].ReportDate.Month() == month {
			return entries
		}
	}

	return nil
}

type ReportEntries []ReportEntry

func (e ReportEntries) String() string {
	projHours := e.ProjectHours()

	groups := e.GroupByWeeks()
	groupStrs := make([]string, 0, len(groups))

	for i := len(groups) - 1; i >= 0; i-- {
		s := groups[i][0].ReportDate.Format("Monday, 2:\n")

		for _, entry := range groups[i] {
			s += "  " + entry.StringShort() + "\n"
		}

		groupStrs = append(groupStrs, s)
	}

	return projHours.String() + "\n\n" + strings.Join(groupStrs, "\n")
}

func (e ReportEntries) Reported(d Date) bool {
	for _, ee := range e {
		if ee.ReportDate.Year() == d.Year() && ee.ReportDate.Month() == d.Month() && ee.ReportDate.Day() == d.Day() {
			return true
		}
	}

	return false
}

func (e ReportEntries) FindLatestForToday(date Date) *ReportEntry {
	sort.Slice(e, func(i, j int) bool {
		return e[i].ReportDate.Add(e[i].EndTime.Sub(e[i].StartTime)).
			After(e[j].ReportDate.Add(e[j].EndTime.Sub(e[j].StartTime)))
	})

	for _, entry := range e {
		if entry.ReportDate.Equal(date) {
			cp := entry

			return &cp
		}
	}

	return nil
}

func (e ReportEntries) Overlaps(entry ReportEntry) ReportEntries {
	var list ReportEntries

	for _, current := range e {
		if !current.ReportDate.Equal(entry.ReportDate) {
			continue
		}

		if current.Overlaps(entry) {
			list = append(list, current)
		}
	}

	return list
}

func (e ReportEntries) ProjectHours() ProjectsHours {
	m := map[Project]ProjectHours{}
	var ph ProjectHours
	var ok bool

	for _, entry := range e {
		if ph, ok = m[entry.Project]; ok {
			ph.Duration += entry.Span
			m[entry.Project] = ph
		} else {
			m[entry.Project] = ProjectHours{Project: entry.Project, Duration: entry.Span}
		}
	}

	projectsHours := make(ProjectsHours, 0, len(m))

	for _, ph = range m {
		projectsHours = append(projectsHours, ph)
	}

	sort.Slice(projectsHours, func(i, j int) bool {
		return projectsHours[i].Duration > projectsHours[j].Duration
	})

	return projectsHours
}

func (e ReportEntries) GroupByWeeks() []ReportEntries {
	var groups []ReportEntries

	//goland:noinspection GoNilness
	for _, entry := range e {
		idx := len(groups) - 1
		if idx < 0 {
			groups = append(groups, ReportEntries{entry})

			continue
		}

		lastGroup := groups[idx]
		lastEntry := lastGroup[len(lastGroup)-1]

		if lastEntry.ReportDate.Equal(entry.ReportDate) {
			groups[idx] = append(groups[idx], entry)
		} else {
			groups = append(groups, ReportEntries{entry})
		}
	}

	return groups
}

func (e ReportEntries) Duration() time.Duration {
	var duration time.Duration

	for _, entry := range e {
		duration += entry.Span
	}

	return duration
}

type ReportEntry struct {
	ID          string
	PublishDate Date
	ReportDate  Date
	Project     Project
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
	ErrNoReportDate   = errors.New("no report date")
	ErrNoAnyTime      = errors.New("no any time")
	ErrTimeOverflow   = errors.New("overflow to the next day")
	ErrStartLargerEnd = errors.New("start time is greater or equal than end time")
	ErrTimeNotRounded = errors.New("time must be rounded to 10 minutes")
	ErrStatusNegative = errors.New("status is less than zero")
	ErrStatusMore100  = errors.New("status is larger than 100")
	ErrStatusNotRound = errors.New("must be rounded to 10")
	ErrBadActivity    = errors.New("bad activity")
	ErrOverlaps       = errors.New("overlaps with existing")
	ErrNoTitle        = errors.New("no title")
	ErrNoMessage      = errors.New("no message")
	ErrNoRange        = errors.New("range must be set")
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

func (e ReportEntry) TestActivity() error {
	activities := []string{
		ActivityEstimate, ActivityDevelopment, ActivityTesting, ActivityBugfixing, ActivityManagement, ActivityAnalysis,
	}

	for _, activity := range activities {
		if activity == e.Activity {
			return nil
		}
	}

	return ErrBadActivity
}

func (e ReportEntry) SetActivity(short string) (ReportEntry, error) {
	activities := []string{
		ActivityEstimate, ActivityDevelopment, ActivityTesting, ActivityBugfixing, ActivityManagement, ActivityAnalysis,
	}

	for _, activity := range activities {
		if !strings.HasPrefix(activity, strings.ToLower(short)) {
			continue
		}

		e.Activity = activity

		return e, nil
	}

	return e, fmt.Errorf("%v: %w", short, ErrBadActivity)
}

func (e ReportEntry) UpdateProjectName(available Projects) (ReportEntry, error) {
	var err error
	if e.Project, err = available.Match(e.Project.Name); err != nil {
		return e, fmt.Errorf("match: %w", err)
	}

	if len(e.Name) == 0 {
		e.Name = e.Project.Name
	}

	return e, nil
}

func (e ReportEntry) AlignTimes(history ReportEntries) (ReportEntry, error) {
	if e.ReportDate.IsZero() {
		return e, fmt.Errorf("zero: %w", ErrNoReportDate)
	}

	if e.IsSpanAndRangeAbsent() {
		return e, fmt.Errorf("no time: %w", ErrNoAnyTime)
	}

	if e.Span == 0 {
		e.Span = e.EndTime.Sub(e.StartTime)
	}

	if e.IsEmptyRange() {
		e.StartTime = time.Date(0, 0, 0, 9, 0, 0, 0, time.UTC)
		if entry := history.FindLatestForToday(e.ReportDate); entry != nil {
			e.StartTime = entry.EndTime
		}

		e.EndTime = e.StartTime.Add(e.Span)

		if e.StartTime.Day() != e.EndTime.Day() {
			return e, fmt.Errorf("overflow: %w", ErrTimeOverflow)
		}
	}

	if entries := history.Overlaps(e); len(entries) > 0 {
		return e, fmt.Errorf("entries: %w", ErrOverlaps)
	}

	return e, nil
}

func (e ReportEntry) String() string {
	return "" +
		"Project:     " + e.Project.Name + "\n" +
		"Report date: " + e.ReportDate.Format("January 2, Monday") + "\n" +
		"Activity:    " + e.Activity + "\n" +
		"Status:      " + strconv.Itoa(e.Status) + "\n" +
		"Time:        " + e.StartTime.Format("15:04") + "-" + e.EndTime.Format("15:04") +
		" (" + e.Span.String() + ")" + "\n" +
		"\n" +
		"Name:        " + e.Name + "\n" +
		"Desc:        " + e.Description
}

func (e ReportEntry) Overlaps(o ReportEntry) bool {
	if (e.StartTime.Equal(o.StartTime) || e.StartTime.Before(o.StartTime)) && e.EndTime.After(o.StartTime) {
		return true
	}

	if e.StartTime.Before(o.EndTime) && (e.EndTime.Equal(o.EndTime) || e.EndTime.After(o.EndTime)) {
		return true
	}

	if e.StartTime.After(o.StartTime) && e.StartTime.Before(o.EndTime) {
		return true
	}

	if e.StartTime.Before(o.StartTime) && e.EndTime.After(o.EndTime) {
		return true
	}

	return false
}

func (e ReportEntry) Test() error {
	if e.ReportDate.IsZero() {
		return ErrNoReportDate
	}

	if len(e.Project.ID) == 0 {
		return ErrProjNotFound
	}

	if err := e.TestActivity(); err != nil {
		return fmt.Errorf("test activity: %w", err)
	}

	if len(e.Name) == 0 {
		return ErrNoTitle
	}

	if len(e.Description) == 0 {
		return ErrNoMessage
	}

	if err := e.TestStatus(); err != nil {
		return fmt.Errorf("test status: %w", err)
	}

	if e.IsEmptyRange() {
		return ErrNoRange
	}

	return nil
}

func (e ReportEntry) URLValues() (url.Values, error) {
	if err := e.Test(); err != nil {
		return nil, fmt.Errorf("test: %w", err)
	}

	return url.Values{
		"report_date":          {e.ReportDate.Format("2006-01-02")},
		"project_id":           {e.Project.ID},
		"activity":             {e.GetActivity()},
		"task_name":            {e.Name},
		"task_desc":            {e.Description},
		"status":               {strconv.Itoa(e.Status)},
		"start_report_hours":   {e.StartTime.Format("15")},
		"start_report_minutes": {e.StartTime.Format("04")},
		"end_report_hours":     {e.EndTime.Format("15")},
		"end_report_minutes":   {e.EndTime.Format("04")},
		"overtime":             {"1"},
	}, nil
}

func (e ReportEntry) GetActivity() string {
	return mapActivityEnum[e.Activity]
}

func (e ReportEntry) IsSame(o ReportEntry) bool {
	return e.ReportDate.Equal(o.ReportDate) &&
		e.Project.Name == o.Project.Name &&
		strings.EqualFold(e.Activity, o.Activity) &&
		e.Name == o.Name &&
		e.Description == o.Description &&
		e.Status == o.Status &&
		e.StartTime.Hour() == o.StartTime.Hour() && e.StartTime.Minute() == o.StartTime.Minute() &&
		e.EndTime.Hour() == o.EndTime.Hour() && e.EndTime.Minute() == o.EndTime.Minute()
}

func (e ReportEntry) StringShort() string {
	return "#" + e.ID + " " + e.Span.String() + " " + e.Name
}

func NewReportEntriesFromHTMLNode(ctx context.Context, doc *html.Node) (ReportEntries, error) {
	_, end := th.RegionTask(ctx, "new report entries from html node")
	defer end()

	expr := `//table[@id="history"]//tbody/tr`
	nodes, err := htmlquery.QueryAll(doc, expr)
	if err != nil {
		return nil, fmt.Errorf("query all '%s': %w", expr, err)
	}

	entries := make(ReportEntries, 0, len(nodes))

	for _, node := range nodes {
		var entry ReportEntry

		iter := []struct {
			s    *string
			expr string
		}{
			{&entry.ID, `./td[1]`},
			{&entry.Project.Name, `./td[4]`},
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

		if strings.Contains(text, ":") {
			text = strings.ReplaceAll(strings.ReplaceAll(text, "h", "m"), ":", "h")
		}

		if entry.Span, err = time.ParseDuration(text); err != nil {
			return nil, fmt.Errorf("parse duration '%s': %w", text, err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
