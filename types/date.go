package types

import (
	"strings"
	"time"
)

type Dates []Date

func (d Dates) String() string {
	if len(d) == 0 {
		return ""
	}

	ss := make([]string, 0, len(d))
	for _, date := range d {
		ss = append(ss, date.String())
	}

	return strings.Join(ss, ", ")
}

type Date struct {
	time.Time
}

func (d Date) Equal(o Date) bool {
	ly, lm, ld := d.Date()
	ry, rm, rd := o.Date()

	return ly == ry && lm == rm && ld == rd
}

func (d Date) AddDate(years, months, day int) Date {
	return Date{d.Time.AddDate(years, months, day)}
}

func (d Date) IsWeekend() bool {
	return d.Weekday() == time.Saturday || d.Weekday() == time.Sunday
}

func (d Date) String() string {
	return d.Time.Format("Monday 2")
}

func ParseDate(layout, value string) (Date, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return Date{}, err
	}

	return Date{t}, nil
}

func Today() Date {
	return Date{time.Now()}
}
