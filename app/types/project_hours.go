package types

import (
	"fmt"
	"strings"
	"time"

	"github.com/jwalton/gchalk"
)

type ProjectsHours []ProjectHours
type ProjectHours struct {
	Project  Project
	Duration time.Duration
}

func (r ProjectsHours) String() string {
	var ss []string

	for _, ph := range r {
		strDuration := gchalk.Green(strings.ReplaceAll(ph.Duration.String(), "0s", ""))
		strProject := gchalk.Magenta(ph.Project.Name)
		ss = append(ss, fmt.Sprintf("%s (%v)", strProject, strDuration))
	}

	return strings.Join(ss, ", ")
}
