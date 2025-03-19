package main

import (
	"fmt"
	"os"

	"github.com/grumpycatyo-collab/go2web/business/cli"
)

func main() {
	options, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cli.PrintHelp()
		os.Exit(1)
	}

	if options.Help {
		cli.PrintHelp()
		os.Exit(0)
	}
}
