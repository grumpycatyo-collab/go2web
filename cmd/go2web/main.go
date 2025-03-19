package main

import (
	"fmt"
	"os"

	"github.com/grumpycatyo-collab/go2web/business/cli"
	"github.com/grumpycatyo-collab/go2web/business/http"
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

	client := http.NewClient()
	if options.URL != "" {
		resp, err := client.Get(options.URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching URL: %v\n", err)
			os.Exit(1)
		}

		fmt.Println(resp.Body)
		os.Exit(0)
	}
}
