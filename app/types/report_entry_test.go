package types_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kudrykv/go-vkpm/app/types"
)

func TestReportEntry_String(t *testing.T) {
	re := types.ReportEntry{
		ID:          "123",
		PublishDate: types.Today(),
		ReportDate:  types.Today().AddDate(0, 0, -1),
		Project:     types.Project{ID: "eggincid", Name: "Egg Inc."},
		Activity:    types.ActivityDevelopment,
		Name:        "doing stuff",
		Description: "a little bit more of a description for doing stuff",
		Status:      70,
		StartTime:   time.Date(0, 0, 0, 10, 0, 0, 0, time.Local),
		EndTime:     time.Date(0, 0, 0, 12, 0, 0, 0, time.Local),
		Span:        2 * time.Hour,
	}

	fmt.Println(re.String())
}
