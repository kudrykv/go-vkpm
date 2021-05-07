package types

import (
	"fmt"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type ReportEntries []ReportEntry

func (e ReportEntries) Reported(d Date) bool {
	for _, ee := range e {
		if ee.ReportDate.Year() == d.Year() && ee.ReportDate.Month() == d.Month() && ee.ReportDate.Day() == d.Day() {
			return true
		}
	}

	return false
}

type ReportEntry struct {
	ID          string
	PublishDate Date
	ReportDate  Date
	ProjectName string
	Activity    string
	Name        string
	Description string
	Status      int
	StartTime   time.Time
	EndTime     time.Time
	Span        time.Duration
}

func NewReportEntriesFromHTMLNode(doc *html.Node) (ReportEntries, error) {
	nodes, err := htmlquery.QueryAll(doc, `//table[@id="history"]//tbody/tr`)
	if err != nil {
		return nil, fmt.Errorf("query all: %w", err)
	}

	entries := make(ReportEntries, 0, len(nodes))

	for _, node := range nodes {
		var entry ReportEntry

		iter := []struct {
			s    *string
			expr string
		}{
			{&entry.ID, `./td[1]`},
			{&entry.ProjectName, `./td[4]`},
			{&entry.Activity, `./td[5]//option[@selected]`},
			{&entry.Name, `./td[6]`},
			{&entry.Description, `./td[7]`},
		}

		for _, kv := range iter {
			if *kv.s, err = getTextFromNode(node, kv.expr); err != nil {
				return nil, fmt.Errorf("get text from node: %w", err)
			}
		}

		iter2 := []struct {
			s      *time.Time
			layout string
			expr   string
		}{

			{&entry.StartTime, `15:04`, `./td[9]`},
			{&entry.EndTime, `15:04`, `./td[10]`},
		}

		for _, kv := range iter2 {
			if *kv.s, err = getTimeFromNode(node, kv.layout, kv.expr); err != nil {
				return nil, fmt.Errorf("get time from node: %w", err)
			}
		}

		iter3 := []struct {
			s      *Date
			layout string
			expr   string
		}{
			{&entry.PublishDate, `2 Jan, Mon 15:04`, `./td[2]`},
			{&entry.ReportDate, `2 Jan, 2006`, `./td[3]`},
		}

		for _, kv := range iter3 {
			if *kv.s, err = getDateFromNode(node, kv.layout, kv.expr); err != nil {
				return nil, fmt.Errorf("get time from node: %w", err)
			}
		}

		if entry.Status, err = getIntFromNode(node, `./td[8]`); err != nil {
			return nil, fmt.Errorf("get int from node: %w", err)
		}

		text, err := getTextFromNode(node, `./td[11]`)
		if err != nil {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		if entry.Span, err = time.ParseDuration(text); err != nil {
			return nil, fmt.Errorf("parse duration: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
