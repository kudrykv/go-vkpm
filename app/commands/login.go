package commands

import (
	"bufio"
	"fmt"
	"os"

	"github.com/kudrykv/go-vkpm/app/commands/before"
	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/kudrykv/go-vkpm/app/printer"
	"github.com/kudrykv/go-vkpm/app/services"
	"github.com/kudrykv/go-vkpm/app/th"
	"github.com/urfave/cli/v2"
	"golang.org/x/term"
)

func Login(p printer.Printer, cfg config.Config, api *services.API) *cli.Command {
	return &cli.Command{
		Name:  "login",
		Usage: "sign in into the system",

		Before: before.IsDomainSet(cfg),

		Action: func(c *cli.Context) error {
			ctx, end := th.RegionTask(c.Context, "login")
			defer end()

			reader := bufio.NewReader(os.Stdin)

			p.Print("username: ")

			username, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("read string: %w", err)
			}

			p.Print("password: ")

			btsPassword, err := term.ReadPassword(0) // 0 is stdin

			p.Println()

			if err != nil {
				return fmt.Errorf("read password: %w", err)
			}

			cfg.Cookies, err = api.Login(ctx, username, string(btsPassword))
			if err != nil {
				return fmt.Errorf("login: %w", err)
			}

			if err = cfg.Write(); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			return nil
		},
	}
}
