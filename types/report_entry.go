package types

import (
	"fmt"
	"time"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
)

type ReportEntries []ReportEntry
type ReportEntry struct {
	ID          string
	PublishDate time.Time
	ReportDate  time.Time
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

	_ = nodes

	return nil, nil
}
