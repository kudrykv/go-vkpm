package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kudrykv/go-vkpm/types"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
)

const (
	fDomain  = "domain"
	fDefProj = "default-project"
)

var (
	errNoConfigInCtx = errors.New("no config in ctx")
	errNoDirInCtx    = errors.New("no dir in ctx")
)

func Config() *cli.Command {
	return &cli.Command{
		Name: "config",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: fDomain},
			&cli.StringFlag{Name: fDefProj},
		},

		Action: func(ctx *cli.Context) error {
			cfg, ok := ctx.Context.Value(types.Cfg).(types.Config)
			if !ok {
				return errNoConfigInCtx
			}

			dir, ok := ctx.Context.Value(types.Dir).(string)
			if !ok {
				return errNoDirInCtx
			}

			if domain := ctx.String(fDomain); len(domain) > 0 {
				cfg.Domain = domain
			}

			if defProj := ctx.String(fDefProj); len(defProj) > 0 {
				cfg.DefaultProject = defProj
			}

			if err := WriteConfig(dir, "config.yml", cfg); err != nil {
				return fmt.Errorf("write config: %w", err)
			}

			return nil
		},
	}
}

func ReadConfig(path, file string) (types.Config, error) {
	bts, err := ioutil.ReadFile(strings.Join([]string{path, file}, string(os.PathSeparator)))
	if err != nil {
		return types.Config{}, fmt.Errorf("read file: %w", err)
	}

	var authConfig types.Config
	if err = yaml.Unmarshal(bts, &authConfig); err != nil {
		return authConfig, fmt.Errorf("unmarshal: %w", err)
	}

	return authConfig, nil
}

func WriteConfig(path, file string, config types.Config) error {
	bts, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if err = ioutil.WriteFile(strings.Join([]string{path, file}, string(os.PathSeparator)), bts, 0600); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
