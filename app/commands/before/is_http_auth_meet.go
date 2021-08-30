package before

import (
	"fmt"

	"github.com/kudrykv/go-vkpm/app/config"
	"github.com/urfave/cli/v2"
)

func IsHTTPAuthMeet(cfg config.Config) func(*cli.Context) error {
	return func(*cli.Context) error {
		if len(cfg.Domain) == 0 {
			return fmt.Errorf("no domain set: %w", ErrNoDomain)
		}

		if cfg.Cookies.IsZero() {
			return fmt.Errorf("sign in first: %w", ErrNoCookies)
		}

		return nil
	}
}
