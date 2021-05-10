package commands

import (
	"fmt"
	"regexp"

	"github.com/kudrykv/go-vkpm/config"
	"github.com/urfave/cli/v2"
)

const (
	flagDomain  = "domain"
	flagDefProj = "defproj"
)

var (
	httpsRegexp = regexp.MustCompile(`^https?://`)
)

func Config(cfg config.Config) *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "update vkpm config",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagDomain, Usage: "domain to use, e.g., domain.com"},
			&cli.StringFlag{Name: flagDefProj, Usage: "report default project"},
		},

		Action: func(ctx *cli.Context) error {
			if domain := ctx.String(flagDomain); len(domain) > 0 {
				if httpsRegexp.MatchString(domain) {
					domain = httpsRegexp.ReplaceAllString(domain, "")
				}

				cfg.Domain = domain
			}

			if defProj := ctx.String(flagDefProj); len(defProj) > 0 {
				cfg.DefaultProject = defProj
			}

			if err := cfg.Write(); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			return nil
		},
	}
}
