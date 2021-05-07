package types

import (
	"errors"
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Holidays []Holiday

func (h Holidays) Holiday(day Date) bool {
	for _, holiday := range h {
		if holiday.Date.Equal(day) {
			return true
		}
	}

	return false
}

func (h Holidays) InMonth(day Date) Holidays {
	var im Holidays

	for _, holiday := range h {
		if holiday.Date.Month() == day.Month() {
			im = append(im, holiday)
		}
	}

	return im
}

func (h Holidays) String() string {
	if len(h) == 0 {
		return ""
	}

	ss := make([]string, 0, len(h))
	for _, holiday := range h {
		ss = append(ss, holiday.String())
	}

	return strings.Join(ss, "\n")
}

type Holiday struct {
	Name string
	Date Date
}

func (h Holiday) String() string {
	return h.Date.String() + " -- " + h.Name
}

func NewHolidaysFromHTMLNode(doc *html.Node) (Holidays, error) {
	nodes, err := htmlquery.QueryAll(doc, `//div[@class="holidays_list"]//tbody/tr`)
	if err != nil {
		return nil, fmt.Errorf("query all: %w", err)
	}

	holidays := make(Holidays, 0, len(nodes))

	var text string

	for _, node := range nodes {
		if text, err = getTextFromNode(node, `./td[2]`); err != nil && !errors.Is(err, ErrNodeNotFound) {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		if len(text) == 0 {
			continue
		}

		var h Holiday
		if h.Date, err = ParseDate(`02 January 2006`, text); err != nil {
			return nil, fmt.Errorf("parse: %w", err)
		}

		if h.Name, err = getTextFromNode(node, `./td[3]`); err != nil {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		holidays = append(holidays, h)
	}

	return holidays, nil
}
