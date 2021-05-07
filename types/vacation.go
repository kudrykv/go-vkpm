package types

import (
	"fmt"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type Vacations []Vacation

func (v Vacations) Vacated(day Date) bool {
	for _, vac := range v {
		if vac.Vacated(day) {
			return true
		}
	}

	return false
}

func (v Vacations) InMonth(day Date) Vacations {
	var vacs Vacations

	for _, vacation := range v {
		if vacation.InMonth(day) {
			vacs = append(vacs, vacation)
		}
	}

	return vacs
}

type Vacation struct {
	ID        string
	Type      string
	StartDate Date
	EndDate   Date
	Span      time.Duration
	Status    string
	Paid      bool
	Note      string
}

func (v Vacation) Vacated(day Date) bool {
	if v.StartDate.Equal(day) {
		return true
	}

	if v.EndDate.IsZero() {
		return false
	}

	cur := v.StartDate
	for i := 0; i < int(v.Span/(time.Duration(24)*time.Hour)); i++ {
		if cur.Equal(day) {
			return true
		}

		cur = cur.AddDate(0, 0, 1)
	}

	return false
}

func (v Vacation) InMonth(day Date) bool {
	if v.StartDate.Month() == day.Month() {
		return true
	}

	if v.EndDate.IsZero() {
		return false
	}

	cur := v.StartDate
	for i := 0; i < int(v.Span/(time.Duration(24)*time.Hour)); i++ {
		if cur.Month() == day.Month() {
			return true
		}

		cur = cur.AddDate(0, 0, 1)
	}

	return false
}

func NewVacationsFromHTMLNode(doc *html.Node) (Vacations, error) {
	nodes, err := htmlquery.QueryAll(doc, `//table[@id="vacations"]/tbody/tr`)
	if err != nil {
		return nil, fmt.Errorf("query all: %w", err)
	}

	vacations := make(Vacations, 0, len(nodes))

	for _, node := range nodes {
		var vac Vacation
		var paidStr string
		var str string
		var i float64

		if vac.Type, err = getTextFromNode(node, `./td[2]`); err != nil {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		if vac.Type == "Vacation Compensation" {
			continue
		}

		strs := []struct {
			s    *string
			expr string
		}{
			{&vac.ID, `./td[1]`},
			{&vac.Status, `./td[6]`},
			{&paidStr, `./td[8]`},
		}

		for _, kv := range strs {
			if *kv.s, err = getTextFromNode(node, kv.expr); err != nil {
				return nil, fmt.Errorf("get text from node: %w", err)
			}
		}

		vac.Paid = paidStr == "Paid"

		if vac.StartDate, err = getDateFromNode(node, `2 January 2006`, `./td[3]`); err != nil {
			return nil, fmt.Errorf("get time from node: %w", err)
		}

		if str, err = getTextFromNode(node, `./td[4]`); err != nil {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		if str != "-" {
			if vac.EndDate, err = ParseDate("2 January 2006", str); err != nil {
				return nil, fmt.Errorf("time parse: %w", err)
			}
		}

		if i, err = getFloat64FromNode(node, `./td[5]`); err != nil {
			return nil, fmt.Errorf("get float64 from node: %w", err)
		}

		vac.Span = time.Duration(i*24) * time.Hour

		vacations = append(vacations, vac)
	}

	return vacations, nil
}
