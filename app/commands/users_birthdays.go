package commands

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/kudrykv/go-vkpm/app/commands/before"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/services"
	"github.com/kudrykv/go-vkpm/app/th"
	"github.com/kudrykv/go-vkpm/app/types"
	"github.com/urfave/cli/v2"
)

const (
	fSortBy = "sort-by"
	fTeam   = "team"
	fName   = "name"

	byID       = "id"
	byName     = "name"
	byTeam     = "team"
	byBirthday = "bday"
)

func UsersBirthdays(cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name: "birthdays",

		Flags: []cli.Flag{
			&cli.StringFlag{Name: fSortBy, Value: byName},
			&cli.StringFlag{Name: fTeam},
			&cli.StringFlag{Name: fName},
		},

		Before: before.IsHTTPAuthMeet(cfg),

		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "users_birthdays")
			defer end()

			sortByItems := strings.Split(c.String(fSortBy), ",")
			for _, item := range sortByItems {
				if item != byID && item != byName && item != byTeam && item != byBirthday {
					return fmt.Errorf("%s: %w", item, errors.New("unknown sort option"))
				}
			}

			persons, err := api.Birthdays(ctx)
			if err != nil {
				return fmt.Errorf("birthdays: %w", err)
			}

			if team := strings.ToLower(c.String(fTeam)); len(team) > 0 {
				var filtered types.Persons

				for _, person := range persons {
					if strings.Contains(strings.ToLower(person.Team), team) {
						filtered = append(filtered, person)
					}
				}

				persons = filtered
			}

			if name := strings.ToLower(c.String(fName)); len(name) > 0 {
				var filtered types.Persons

				for _, person := range persons {
					if strings.Contains(strings.ToLower(person.Name), name) {
						filtered = append(filtered, person)
					}
				}

				persons = filtered
			}

			sort.Slice(persons, sortingPersons(sortByItems, persons))

			_, _ = fmt.Fprintln(c.App.Writer, persons)

			return nil
		},
	}
}

func sortingPersons(sortItems []string, persons types.Persons) func(i, j int) bool {
	return func(i, j int) bool {
		left := ""
		right := ""

		for _, item := range sortItems {
			switch item {
			case byID:
				left += fmt.Sprintf("%06d", persons[i].ID)
				right += fmt.Sprintf("%06d", persons[j].ID)
			case byName:
				left += persons[i].Name
				right += persons[j].Name
			case byTeam:
				left += persons[i].Team
				right += persons[j].Team
			case byBirthday:
				left += persons[i].Birthday.Format(time.RFC3339)
				right += persons[j].Birthday.Format(time.RFC3339)
			}
		}

		return left < right
	}
}
