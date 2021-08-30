package types

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/go-vkpm/app/th"
	"github.com/olekukonko/tablewriter"
	"golang.org/x/net/html"
)

type Persons []Person
type Person struct {
	ID             int
	URL            string
	Name           string
	Email          string
	Skype          string
	Grade          string
	Birthday       time.Time
	Team           string
	EnglishLevel   string
	EnglishDetails string
	PhotoURL       string
}

func (p Person) Row() []string {
	return []string{strconv.Itoa(p.ID), p.Name, p.Team, p.Birthday.Format("Jan _2")}
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

	for _, node := range nodes {
		expr = `./td[1]/a`
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

func NewPersonUserProfileFromHTMLNode(_ context.Context, doc *html.Node) (Person, error) {
	person := Person{}
	var err error

	if person.Email, err = getTextFromNode(doc, `//td[contains(., 'Email')]/following-sibling::td`); err != nil {
		return person, fmt.Errorf("query email: %w", err)
	}

	if person.Skype, err = getTextFromNode(doc, `//td[contains(., 'Skype')]/following-sibling::td`); err != nil {
		return person, fmt.Errorf("query skype: %w", err)
	}

	if person.Team, err = getTextFromNode(doc, `//td[contains(., 'Team')]/following-sibling::td`); err != nil {
		return person, fmt.Errorf("query team: %w", err)
	}

	if person.Grade, err = getTextFromNode(doc, `//td[contains(., 'Grade')]/following-sibling::td`); err != nil {
		return person, fmt.Errorf("query grade: %w", err)
	}

	person.EnglishLevel, err = getTextFromNode(doc, `//td[contains(., 'English Level')]/following-sibling::td`)
	if err != nil {
		return person, fmt.Errorf("query english level: %w", err)
	}

	person.EnglishDetails, err = getTextFromNode(doc, `//td[contains(., 'English Details')]/following-sibling::td`)
	if err != nil {
		return person, fmt.Errorf("query english details: %w", err)
	}

	query, err := htmlquery.Query(doc, `//td[contains(., 'Photo')]/following-sibling::td//img`)
	if err != nil {
		return person, fmt.Errorf("query photo src: %w", err)
	}

	for _, attr := range query.Attr {
		if attr.Key == "src" {
			person.PhotoURL = attr.Val
		}
	}

	return person, nil
}

func (p Persons) String() string {
	builder := strings.Builder{}
	table := tablewriter.NewWriter(&builder)

	table.SetHeader([]string{"ID", "Name", "Team", "Birthday"})

	for _, person := range p {
		table.Append(person.Row())
	}

	table.Render()

	return builder.String()
}
