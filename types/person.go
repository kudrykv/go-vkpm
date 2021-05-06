package types

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
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

func NewPersonsFromHTMLNode(doc *html.Node) (Persons, error) {
	var (
		persons Persons
		err     error
		ptr     *html.Node
	)

	nodes, err := htmlquery.QueryAll(doc, `//*[@id="dashboard_birthdays_block"]//tbody/tr`)
	if err != nil {
		return persons, fmt.Errorf("query all: %w", err)
	}

	persons = make(Persons, 0, len(nodes))

	for _, node := range nodes {
		person := Person{}

		if ptr, err = htmlquery.Query(node, `./td[1]/a`); err != nil {
			return persons, fmt.Errorf("query name: %w", err)
		}

		for _, attr := range ptr.Attr {
			if attr.Key != "href" {
				continue
			}

			person.URL = attr.Val
			if index := strings.LastIndex(person.URL, "/"); index >= 0 {
				if person.ID, err = strconv.Atoi(person.URL[index+1:]); err != nil {
					return persons, fmt.Errorf("atoi: %w", err)
				}
			}
		}

		person.Name = strings.TrimSpace(ptr.FirstChild.Data)

		if ptr, err = htmlquery.Query(node, `./td[2]`); err != nil {
			return persons, fmt.Errorf("query bday: %w", err)
		}

		if person.Birthday, err = time.Parse("02 January", strings.TrimSpace(ptr.FirstChild.Data)); err != nil {
			return persons, fmt.Errorf("parse bday: %w", err)
		}

		if ptr, err = htmlquery.Query(node, `./td[3]`); err != nil {
			return persons, fmt.Errorf("query team: %w", err)
		}

		person.Team = strings.TrimSpace(ptr.FirstChild.Data)

		persons = append(persons, person)
	}

	return persons, nil
}
