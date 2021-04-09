package commands

import (
	"context"
	"fmt"

	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
)

func Login() *cli.Command {
	return &cli.Command{
		Name: "login",
		Before: func(ctx *cli.Context) error {
			dir, ok := ctx.Context.Value(types.Dir).(string)
			if !ok {
				return errNoDirInCtx
			}

			config, err := ReadConfig(dir, "config.yml")
			if err != nil {
				return fmt.Errorf("read config: %w", err)
			}

			ctx.Context = context.WithValue(ctx.Context, types.Cfg, config)

			return nil
		},
		Action: func(ctx *cli.Context) error {
			return nil
		},
	}
}
