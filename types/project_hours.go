package types

import (
	"fmt"
	"strings"
	"time"
)

type ProjectsHours []ProjectHours
type ProjectHours struct {
	Project  Project
	Duration time.Duration
}

func (r ProjectsHours) String() string {
	var ss []string

	for _, ph := range r {
		ss = append(ss, fmt.Sprintf("%s (%v)", ph.Project.Name, ph.Duration))
	}

	return strings.Join(ss, ", ")
}
