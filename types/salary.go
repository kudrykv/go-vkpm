package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/jwalton/gchalk"
	"golang.org/x/net/html"
)

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
	Month              int
	Total              float64
	Paid               float64
}

func (s Salary) StringTotalPaid() string {
	format := time.Date(s.Year, time.Month(s.Month), 0, 0, 0, 0, 0, time.UTC).Format("January, 2006")

	var out string
	if s.Paid == 0 {
		expected := gchalk.Green("$" + strconv.FormatFloat(s.ExpectedSalary, 'f', 2, 64))
		out = "expected " + expected
	} else {
		paid := gchalk.Green("$" + strconv.FormatFloat(s.Paid, 'f', 2, 64))
		out = "got " + paid
	}

	return gchalk.White(format) + ": " + out
}

func (s Salary) StringHoursReport() string {
	return "Reported: " + f2s(s.HoursByCurrDay) + " of " + f2s(s.WorkingDaysInMonth)
}

func f2s(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

var (
	ErrNodeNotFound = errors.New("node not found")
	ErrNumNotFound  = errors.New("number not found")
	ErrBadTotalPaid = errors.New("bad total / paid")

	numRegex = regexp.MustCompile(`\d+(?:\.\d+)?`)
)

func NewSalaryFromHTMLNode(doc *html.Node, year int, month int) (Salary, error) {
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
		if *kv.f, err = getNumFromNode(doc, kv.expr); err != nil {
			return salary, fmt.Errorf("get num from node: %w", err)
		}
	}

	totalAndPaid, err := getTextFromNode(doc, `//td[. = "Total / Paid"]/following-sibling::td`)
	if err != nil {
		return salary, fmt.Errorf("get text from node: %w", err)
	}

	strings := numRegex.FindAllString(totalAndPaid, -1)
	if len(strings) != 2 {
		return salary, fmt.Errorf(totalAndPaid+": %w", ErrBadTotalPaid)
	}

	totalStr, paidStr := strings[0], strings[1]

	if salary.Total, err = strconv.ParseFloat(totalStr, 10); err != nil {
		return salary, fmt.Errorf("parse float: %w", err)
	}

	if salary.Paid, err = strconv.ParseFloat(paidStr, 10); err != nil {
		return salary, fmt.Errorf("parse float: %w", err)
	}

	return salary, nil
}

func getNumFromNode(doc *html.Node, expr string) (float64, error) {
	text, err := getTextFromNode(doc, expr)
	if err != nil {
		return 0, fmt.Errorf("get text from node: %w", err)
	}

	num, err := getNumFromString(text)
	if err != nil {
		return 0, fmt.Errorf("get num from string: %w", err)
	}

	return num, nil
}

func getTextFromNode(doc *html.Node, expr string) (string, error) {
	node, err := htmlquery.Query(doc, expr)
	if err != nil {
		return "", fmt.Errorf("query rate per hour: %w", err)
	}

	if node == nil {
		return "", fmt.Errorf("rate per hour: %w", ErrNodeNotFound)
	}

	return node.FirstChild.Data, nil
}

func getNumFromString(str string) (float64, error) {
	numStr := numRegex.FindString(str)
	if len(numStr) == 0 {
		return 0, fmt.Errorf("search in '%v': %w", str, ErrNumNotFound)
	}

	num, err := strconv.ParseFloat(numStr, 10)
	if err != nil {
		return 0, fmt.Errorf("parse float: %w", err)
	}

	return num, nil
}
