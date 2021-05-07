package types

import "time"

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
