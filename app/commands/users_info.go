package commands

import (
	"bytes"
	"fmt"
	"image/color"
	"strconv"

	"github.com/eliukblau/pixterm/pkg/ansimage"
	"github.com/kudrykv/go-vkpm/app/commands/before"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/services"
	"github.com/kudrykv/go-vkpm/app/types"
	"github.com/urfave/cli/v2"
)

const (
	fID      = "id"
	fWithPic = "with-pic"
	fXSize   = "x"
	fYSize   = "y"
)

func UsersInfo(cfg config.Config, api *services.API) *cli.Command {
	var (
		user types.Person
		id   int
		err  error
	)

	return &cli.Command{
		Name: "info",

		Flags: []cli.Flag{
			&cli.StringFlag{Name: fID, Required: true},
			&cli.BoolFlag{Name: fWithPic},
			&cli.IntFlag{Name: fXSize, Value: 80},
			&cli.IntFlag{Name: fYSize, Value: 80},
		},

		Before: before.IsHTTPAuthMeet(cfg),

		Action: func(c *cli.Context) error {
			if id, err = strconv.Atoi(c.String(fID)); err != nil {
				return fmt.Errorf("%s: %w", c.String(fID), err)
			}

			user, err = api.UserInfo(c.Context, id)
			if err != nil {
				return fmt.Errorf("user info: %w", err)
			}

			_, _ = fmt.Fprintln(c.App.Writer, user)

			if c.Bool(fWithPic) {
				bts, err := api.GetPicture(c.Context, user.PhotoURL)
				if err != nil {
					return fmt.Errorf("get picture: %w", err)
				}

				x := c.Int(fXSize)
				y := c.Int(fYSize)
				reader := bytes.NewReader(bts)
				bg := color.Black
				ansImage, err := ansimage.NewScaledFromReader(reader, y, x, bg, ansimage.ScaleModeFit, ansimage.NoDithering)

				if err != nil {
					return fmt.Errorf("image from url: %w", err)
				}

				_, _ = fmt.Fprintln(c.App.Writer, ansImage.Render())
			}

			return nil
		},
	}
}
