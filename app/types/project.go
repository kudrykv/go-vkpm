package types

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/kudrykv/go-vkpm/app/th"
	"golang.org/x/net/html"
)

type Projects []Project

var (
	ErrTooPermissive = errors.New("too permissive")
	ErrProjNotFound  = errors.New("project not found")
)

func (p Projects) Match(name string) (Project, error) {
	var matched Projects

	for _, project := range p {
		if idx := strings.Index(strings.ToLower(project.Name), strings.ToLower(name)); idx >= 0 {
			matched = append(matched, project)
		}
	}

	if len(matched) == 0 {
		return Project{}, fmt.Errorf("lookup %v: %w", name, ErrProjNotFound)
	}

	if len(matched) > 1 {
		return Project{}, fmt.Errorf("found multiple %s: %w", matched, ErrTooPermissive)
	}

	return matched[0], nil
}

func (p Projects) String() string {
	if len(p) == 0 {
		return ""
	}

	ss := make([]string, 0, len(p))

	for _, project := range p {
		ss = append(ss, project.Name)
	}

	return strings.Join(ss, ", ")
}

type Project struct {
	ID   string
	Name string
}

func NewProjectsFromHTMLNode(ctx context.Context, doc *html.Node) (Projects, error) {
	_, end := th.RegionTask(ctx, "new projects from html node")
	defer end()

	expr := `//select[@id="id_project"]/option`
	nodes, err := htmlquery.QueryAll(doc, expr)
	if err != nil {
		return nil, fmt.Errorf("query all '%s': %w", expr, err)
	}

	projects := make(Projects, 0, len(nodes))

	for _, node := range nodes {
		var project Project

		if project.Name, err = getTextFromNode(node, `.`); err != nil {
			return nil, fmt.Errorf("get text from node: %w", err)
		}

		for _, attr := range node.Attr {
			if attr.Key == "value" {
				project.ID = attr.Val
			}
		}

		projects = append(projects, project)
	}

	return projects, nil
}
