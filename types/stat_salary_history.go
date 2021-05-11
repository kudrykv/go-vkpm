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
		salary := s.Salaries.At(s.Year, i)
		_ = salary
	}

	return b.String()
}
