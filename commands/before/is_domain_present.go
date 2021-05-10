package before

import (
	"errors"
	"fmt"

	"github.com/kudrykv/go-vkpm/config"
	"github.com/urfave/cli/v2"
)

var (
	ErrNoDomain = errors.New("empty domain")
)

func IsDomainPresent(cfg config.Config) func(*cli.Context) error {
	return func(*cli.Context) error {
		if len(cfg.Domain) == 0 {
			return fmt.Errorf("domain must be present: %w", ErrNoDomain)
		}

		return nil
	}
}
