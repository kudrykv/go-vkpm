package types_test

import (
	"bytes"
	"testing"

	"github.com/kudrykv/go-vkpm/types"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/html"
)

func TestNewHolidaysFromHTMLNode(t *testing.T) {
	Convey("NewHolidaysFromHTMLNode", t, func() {
		node, err := html.Parse(bytes.NewReader([]byte(holidaysHTML)))
		So(err, ShouldBeNil)

		expected := types.Holidays{
			{Name: "Labour Day", Date: atDate(holidayLayout, "03 May 2021")},
			{Name: "Easter", Date: atDate(holidayLayout, "04 May 2021")},
			{Name: "Fun Day", Date: atDate(holidayLayout, "10 May 2021")},
		}

		holidays, err := types.NewHolidaysFromHTMLNode(node)
		So(err, ShouldBeNil)
		So(holidays, ShouldResemble, expected)
	})
}

const holidaysHTML = `
<html><head></head>
<body>
<div class="holidays_list">
<table>
<tbody>
<tr> <td>May</td> <td></td> <td></td> </tr>
<tr> <td></td> <td>03 May 2021</td> <td>Labour Day</td> </tr>
<tr> <td></td> <td>04 May 2021</td> <td>Easter</td> </tr>
<tr> <td></td> <td>10 May 2021</td> <td>Fun Day</td> </tr>
</tbody>
</table>
</div>
</body></html>
`

const holidayLayout = `02 January 2006`

func atDate(layout, value string) types.Date {
	date, err := types.ParseDate(layout, value)
	if err != nil {
		panic(err)
	}

	return date
}
