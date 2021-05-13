package types_test

import (
	"testing"
	"time"

	"github.com/kudrykv/vkpm/types"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDate_Equal(t *testing.T) {
	Convey("Equal", t, func() {
		p1 := types.Today()
		p2 := types.Date{Time: p1.Round(time.Second)}

		So(p1.Equal(p2), ShouldBeTrue)
	})
}
