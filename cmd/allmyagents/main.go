package main

import (
	"os"

	"github.com/n7ptd2xr8c-cell/allmyagents/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}
