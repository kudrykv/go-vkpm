package types

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/jwalton/gchalk"
	"golang.org/x/net/html"
)

type Salaries []Salary

func (s Salaries) Paid() float64 {
	var paid float64

	for _, salary := range s {
		paid += salary.Paid
	}

	return paid
}

func (s Salaries) Expected() float64 {
	var expected float64

	for _, salary := range s {
		expected += salary.ExpectedSalary
	}

	return expected
}

func (s Salaries) At(year int, month time.Month) Salary {
	for _, salary := range s {
		if salary.Year == year && salary.Month == month {
			return salary
		}
	}

	return Salary{}
}

type Salary struct {
	RatePerHour        float64
	Rate               float64
	HoursByCurrDay     float64
	DollarsByCurrDay   float64
	ExpectedSalary     float64
	VacationHours      float64
	VacationDollars    float64
	OvertimeHours      float64
	OvertimeDollars    float64
	BonusDollars       float64
	WorkingDaysInMonth float64
	Year               int
	Month              time.Month
	Total              float64
	Paid               float64
}

func (s Salary) StringTotalPaid() string {
	format := time.Date(s.Year, time.Month(s.Month), 1, 0, 0, 0, 0, time.UTC).Format("January, 2006")

	return gchalk.White(format) + ": " + s.StringTotalPaidShort()
}

func (s Salary) StringTotalPaidShort() string {
	var out string
	if s.Paid == 0 {
		expected := gchalk.Green("$" + strconv.FormatFloat(s.ExpectedSalary, 'f', 2, 64))
		out = "expected " + expected
	} else {
		paid := gchalk.Green("$" + strconv.FormatFloat(s.Paid, 'f', 2, 64))
		out = "got " + paid
	}

	return out
}

func (s Salary) StringHoursReport() string {
	return "Reported: " + f2s(s.HoursByCurrDay) + " of " + f2s(s.WorkingDaysInMonth)
}

func f2s(f float64, precision ...int) string {
	prec := 2
	if len(precision) > 0 {
		prec = precision[0]
	}

	return strconv.FormatFloat(f, 'f', prec, 64)
}

var (
	ErrBadTotalPaid = errors.New("bad total / paid")

	floatRegex = regexp.MustCompile(`\d+(?:\.\d+)?`)
	intRegex   = regexp.MustCompile(`\d+`)
)

func NewSalaryFromHTMLNode(ctx context.Context, doc *html.Node, year int, month time.Month) (Salary, error) {
	_, t := trace.NewTask(ctx, "salary struct from html node")
	defer t.End()

	var (
		salary = Salary{Year: year, Month: month}
		err    error
	)

	iter := []struct {
		f    *float64
		expr string
	}{
		{&salary.RatePerHour, `//td[contains(., "rate per hour:")]`},
		{&salary.Rate, `//td[contains(., "rate:")]`},
		{&salary.WorkingDaysInMonth, `//td[. = "Working days in month:"]/following-sibling::td[1]`},
		{&salary.HoursByCurrDay, `//td[. = "Hours By Current Day"]/following-sibling::td[1]`},
		{&salary.DollarsByCurrDay, `//td[. = "Hours By Current Day"]/following-sibling::td[2]`},
		{&salary.ExpectedSalary, `//td[. = "Expected Salary"]/following-sibling::td[2]`},
		{&salary.VacationHours, `//td[. = "Vacations"]/following-sibling::td[1]`},
		{&salary.VacationDollars, `//td[. = "Vacations"]/following-sibling::td[2]`},
		{&salary.OvertimeHours, `//td[. = "Overtimes"]/following-sibling::td[1]`},
		{&salary.OvertimeDollars, `//td[. = "Overtimes"]/following-sibling::td[2]`},
		{&salary.BonusDollars, `//td[. = "Bonuses"]/following-sibling::td[2]`},
	}

	for _, kv := range iter {
		if *kv.f, err = getFloaty64FromNode(doc, kv.expr); err != nil {
			return salary, fmt.Errorf("get num from node: %w", err)
		}
	}

	totalAndPaid, err := getTextFromNode(doc, `//td[. = "Total / Paid"]/following-sibling::td`)
	if err != nil {
		return salary, fmt.Errorf("get text from node: %w", err)
	}

	strs := floatRegex.FindAllString(totalAndPaid, -1)
	if len(strs) != 2 {
		return salary, fmt.Errorf(totalAndPaid+": %w", ErrBadTotalPaid)
	}

	totalStr, paidStr := strs[0], strs[1]

	if salary.Total, err = strconv.ParseFloat(totalStr, 10); err != nil {
		return salary, fmt.Errorf("parse float: %w", err)
	}

	if salary.Paid, err = strconv.ParseFloat(paidStr, 10); err != nil {
		return salary, fmt.Errorf("parse float: %w", err)
	}

	return salary, nil
}

func getFloaty64FromNode(doc *html.Node, expr string) (float64, error) {
	text, err := getTextFromNode(doc, expr)
	if err != nil {
		return 0, fmt.Errorf("get text from node: %w", err)
	}

	num, err := getFloatFromDirtyString(text)
	if err != nil {
		return 0, fmt.Errorf("get num from string: %w", err)
	}

	return num, nil
}

func getIntyFromNode(doc *html.Node, expr string) (int, error) {
	text, err := getTextFromNode(doc, expr)
	if err != nil {
		return 0, fmt.Errorf("get text from node: %w", err)
	}

	if len(text) == 0 {
		return 0, nil
	}

	num, err := getIntFromDirtyString(text)
	if err != nil {
		return 0, fmt.Errorf("get int from dirty string: %w", err)
	}

	return num, nil
}

func getTextFromNode(doc *html.Node, expr string) (string, error) {
	node, err := htmlquery.Query(doc, expr)
	if err != nil {
		return "", fmt.Errorf("query '%s': %w", expr, err)
	}

	if node == nil {
		return "", nil
	}

	if node.FirstChild == nil {
		return "", nil
	}

	return strings.TrimSpace(node.FirstChild.Data), nil
}

func getTimeFromNode(doc *html.Node, layout, expr string) (time.Time, error) {
	text, err := getTextFromNode(doc, expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("get text from node: %w", err)
	}

	moment, err := time.Parse(layout, text)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse time: %w", err)
	}

	return moment, nil
}

func getDateFromNode(doc *html.Node, layout, expr string) (Date, error) {
	text, err := getTextFromNode(doc, expr)
	if err != nil {
		return Date{}, fmt.Errorf("get text from node: %w", err)
	}

	moment, err := time.Parse(layout, text)
	if err != nil {
		return Date{}, fmt.Errorf("parse time: %w", err)
	}

	return Date{moment}, nil
}

func getFloatFromDirtyString(str string) (float64, error) {
	numStr := floatRegex.FindString(str)
	if len(numStr) == 0 {
		return 0, nil
	}

	num, err := strconv.ParseFloat(numStr, 10)
	if err != nil {
		return 0, fmt.Errorf("parse float: %w", err)
	}

	return num, nil
}

func getIntFromDirtyString(str string) (int, error) {
	intStr := intRegex.FindString(str)

	atoi, err := strconv.Atoi(intStr)
	if err != nil {
		return 0, fmt.Errorf("atoi '%s': %w", str, err)
	}

	return atoi, nil
}
