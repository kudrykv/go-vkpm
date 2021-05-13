package commands

import (
	"fmt"
	"regexp"

	"github.com/kudrykv/vkpm/config"
	"github.com/kudrykv/vkpm/th"
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
		Usage: "change config",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: flagDomain, Usage: "domain to use, e.g., domain.com"},
			&cli.StringFlag{Name: flagDefProj, Usage: "report time for the given project if none specifed in report"},
		},

		Action: func(c *cli.Context) error {
			_, end := th.RegionTask(c.Context, "config")
			defer end()

			if domain := c.String(flagDomain); len(domain) > 0 {
				if httpsRegexp.MatchString(domain) {
					domain = httpsRegexp.ReplaceAllString(domain, "")
				}

				cfg.Domain = domain
			}

			if defProj := c.String(flagDefProj); len(defProj) > 0 {
				cfg.DefaultProject = defProj
			}

			if err := cfg.Write(); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			return nil
		},
	}
}
