package types

import (
	"strconv"
	"strings"
	"time"

	"github.com/jwalton/gchalk"
)

type StatSalaryHistory struct {
	Year       int
	StartMonth time.Month
	EndMonth   time.Month
	Histories  GroupedEntries
	Salaries   Salaries
}

func (s StatSalaryHistory) String() string {
	b := strings.Builder{}

	b.WriteString(strconv.Itoa(s.Year) + " report\n\n")

	money := []string{"Got " + gchalk.Green("$"+f2s(s.Salaries.Paid()))}

	if expected := s.Salaries.Expected(); expected > 0 {
		money = append(money, "expecting "+gchalk.Green("$"+f2s(expected)))
	}

	b.WriteString(strings.Join(money, ", ") + "\n\n")

	for i := s.StartMonth; i <= s.EndMonth; i++ {
		history := s.Histories.At(s.Year, i)
		if len(history) == 0 {
			continue
		}

		salary := s.Salaries.At(s.Year, i)

		b.WriteString(gchalk.White(salary.Month.String()) + ": " + salary.StringTotalPaidShort())
		b.WriteString(", spent " + history.Duration().String() + " in " + history.ProjectHours().String())
		b.WriteString("\n")
	}

	return b.String()
}
