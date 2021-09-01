package commands

import (
	"bytes"
	"errors"
	"fmt"
	"image/color"
	"strconv"
	"strings"

	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/kudrykv/go-vkpm/app/commands/before"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/services"
	"github.com/kudrykv/go-vkpm/app/types"
	"github.com/urfave/cli/v2"
)

const (
	fWithPic = "with-pic"
	fDim     = "dim"
)

func UsersInfo(cfg config.Config, api *services.API) *cli.Command {
	var (
		id      int
		search  string
		person  types.Person
		persons types.Persons
		err     error
	)

	return &cli.Command{
		Name:  "info",
		Usage: "get more details about the user",

		Flags: []cli.Flag{
			&cli.BoolFlag{Name: fWithPic, Usage: "print user image"},
			&cli.IntFlag{Name: fDim, Value: 80},
		},

		Before: before.IsHTTPAuthMeet(cfg),

		Action: func(c *cli.Context) error {
			if search = strings.Join(c.Args().Slice(), " "); len(search) == 0 {
				return errors.New("specify user id or name")
			}

			if id, err = strconv.Atoi(search); err != nil {
				if persons, err = api.Birthdays(c.Context); err != nil {
					return fmt.Errorf("birthdays %s: %w", search, err)
				}

				persons = persons.Filter(types.PersonsFilter{Type: types.ByName, Value: search})

				if len(persons) == 0 {
					return fmt.Errorf("%s: %w", search, errors.New("no one found"))
				}

				if len(persons) > 1 {
					_, _ = fmt.Fprintln(c.App.Writer, persons)

					return fmt.Errorf("%s: %w", search, errors.New("found multiple users"))
				}

				id = persons[0].ID
			}

			person, err = api.PersonInfo(c.Context, id)
			if err != nil {
				return fmt.Errorf("person info: %w", err)
			}

			_, _ = fmt.Fprintln(c.App.Writer, person)

			if c.Bool(fWithPic) {
				bts, err := api.GetPicture(c.Context, person.PhotoURL)
				if err != nil {
					return fmt.Errorf("get picture: %w", err)
				}

				dim := c.Int(fDim)
				reader := bytes.NewReader(bts)
				bg := color.Black
				ansImage, err := ansimage.NewScaledFromReader(reader, dim, dim, bg, ansimage.ScaleModeFit, ansimage.NoDithering)

				if err != nil {
					return fmt.Errorf("image from url: %w", err)
				}

				_, _ = fmt.Fprintln(c.App.Writer, ansImage.Render())
			}

			return nil
		},
	}
}
