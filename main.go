package main

import (
	"fmt"
	"os"

	"github.com/kudrykv/go-vkpm/commands"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name: "vkpm",
		Commands: []*cli.Command{
			commands.Config(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("error:", err) //nolint:forbidigo
	}
}
