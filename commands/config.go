package commands

import (
	"fmt"

	"github.com/kudrykv/go-vkpm/config"
	"github.com/urfave/cli/v2"
)

const (
	fDomain  = "domain"
	fDefProj = "default-project"
)

func Config(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name: "config",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: fDomain},
			&cli.StringFlag{Name: fDefProj},
		},

		Action: func(ctx *cli.Context) error {
			dir, err := config.EnsureDir()
			if err != nil {
				return fmt.Errorf("ensure config dir: %w", err)
			}

			if domain := ctx.String(fDomain); len(domain) > 0 {
				cfg.Domain = domain
			}

			if defProj := ctx.String(fDefProj); len(defProj) > 0 {
				cfg.DefaultProject = defProj
			}

			if err := config.Write(dir, config.Filename, cfg); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			return nil
		},
	}
}
