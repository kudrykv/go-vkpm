package types

import (
	"context"
	"fmt"
	"regexp"
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
	Status         string
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

var regexMultiSpace = regexp.MustCompile(`\s+`)

func (p Person) String() string {
	lines := []string{p.Name}

	if len(p.Status) > 0 {
		lines[0] += " (" + strings.TrimRight(regexMultiSpace.ReplaceAllString(p.Status, " "), ",") + ")"
	}

	lines = append(lines, "Team: "+p.Team, "Email: "+p.Email, "Skype: "+p.Skype)

	if len(p.Grade) > 0 {
		lines = append(lines, "Grade: "+p.Grade)
	}

	if len(p.EnglishLevel) > 0 {
		lines = append(lines, "English level: "+p.EnglishLevel)
	}

	if len(p.EnglishDetails) > 0 {
		lines = append(lines, "English details: "+p.EnglishDetails)
	}

	return strings.Join(lines, "\n")
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

	if person.Name, err = getTextFromNode(doc, `//h1`); err != nil {
		return person, fmt.Errorf("query name: %w", err)
	} else {

	}

	if person.Status, err = getTextFromNode(doc, `//div[@class="status"]`); err != nil {
		return person, fmt.Errorf("query status: %w", err)
	}

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
