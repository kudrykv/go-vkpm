package types

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/vkpm/th"
	"golang.org/x/net/html"
)

type Persons []Person
type Person struct {
	ID       int
	URL      string
	Name     string
	Birthday time.Time
	Team     string
}

func NewPersonsFromHTMLNode(ctx context.Context, doc *html.Node) (Persons, error) {
	_, end := th.RegionTask(ctx, "new persons from html node")
	defer end()

	var (
		persons Persons
		err     error
		ptr     *html.Node
	)

	expr := `//*[@id="dashboard_birthdays_block"]//tbody/tr`
	nodes, err := htmlquery.QueryAll(doc, expr)
	if err != nil {
		return persons, fmt.Errorf("query all by '%s': %w", expr, err)
	}

	persons = make(Persons, 0, len(nodes))

	expr = `./td[1]/a`

	for _, node := range nodes {
		person := Person{}

		if ptr, err = htmlquery.Query(node, expr); err != nil {
			return persons, fmt.Errorf("query name '%s': %w", expr, err)
		}

		for _, attr := range ptr.Attr {
			if attr.Key != "href" {
				continue
			}

			person.URL = attr.Val
			if index := strings.LastIndex(person.URL, "/"); index >= 0 {
				if person.ID, err = strconv.Atoi(person.URL[index+1:]); err != nil {
					return persons, fmt.Errorf("atoi '%s': %w", person.URL[index+1:], err)
				}
			}
		}

		person.Name = strings.TrimSpace(ptr.FirstChild.Data)

		expr = `./td[2]`
		if ptr, err = htmlquery.Query(node, expr); err != nil {
			return persons, fmt.Errorf("query bday '%s': %w", expr, err)
		}

		if person.Birthday, err = time.Parse("02 January", strings.TrimSpace(ptr.FirstChild.Data)); err != nil {
			return persons, fmt.Errorf("parse bday '%s': %w", ptr.FirstChild.Data, err)
		}

		expr = `./td[3]`
		if ptr, err = htmlquery.Query(node, expr); err != nil {
			return persons, fmt.Errorf("query team '%s': %w", expr, err)
		}

		person.Team = strings.TrimSpace(ptr.FirstChild.Data)

		persons = append(persons, person)
	}

	return persons, nil
}
