package types

import (
	"fmt"
	"time"

	"github.com/jwalton/gchalk"
)

type MonthInfo struct {
	salary    Salary
	vacations Vacations
	holidays  Holidays
	moment    Date
	history   ReportEntries
}

func NewMonthInfo(date Date, s Salary, v Vacations, h Holidays, history ReportEntries) MonthInfo {
	return MonthInfo{
		moment:    date,
		salary:    s,
		vacations: v,
		holidays:  h,
		history:   history,
	}
}

func (m MonthInfo) String() string {
	him := m.salary.WorkingDaysInMonth * 8
	wd := len(m.workingDays())
	dead, nrep := m.needReporting()

	s := fmt.Sprintf("Hours in month: %.f (%.f days)\n", him, m.salary.WorkingDaysInMonth)
	s += fmt.Sprintf("Reported as of today: %.1f / %d\n", m.salary.HoursByCurrDay, wd*8)

	if len(dead) > 0 {
		s += gchalk.Gray("Missed reporting for " + dead.String() + "\n")
	}

	if len(nrep) > 0 {
		s += gchalk.Yellow("Need to report for " + nrep.String() + "\n")
	}

	if him := m.holidays.InMonth(m.moment); len(him) > 0 {
		s += fmt.Sprint("\nHolidays:\n", him, "\n")
	}

	if vim := m.vacations.InMonth(m.moment); len(vim) > 0 {
		s += fmt.Sprint("\nVacations:\n", vim, "\n")
	}

	return s
}

func (m MonthInfo) workingDays() []Date {
	days := make([]Date, 0, m.moment.Day())

	cursor := m.moment.AddDate(0, 0, -m.moment.Day())
	for i := 0; i < m.moment.Day(); i++ {
		cursor = cursor.AddDate(0, 0, 1)

		if cursor.IsWeekend() {
			continue
		}

		if m.vacations.Vacated(cursor) {
			continue
		}

		if m.holidays.Holiday(cursor) {
			continue
		}

		days = append(days, cursor)
	}

	return days
}

func (m MonthInfo) needReporting() (Dates, Dates) {
	var need Dates

	for _, day := range m.workingDays() {
		if m.history.Reported(day) {
			continue
		}

		need = append(need, day)
	}

	deadIdx := 0
	day := time.Now().Day()

	for i := len(need) - 1; i >= 0; i-- {
		date := need[i]

		if day-date.Day() > 3 {
			deadIdx = i + 1

			break
		}
	}

	return need[:deadIdx], need[deadIdx:]
}
