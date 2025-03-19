package main

import (
	"fmt"
	"os"

	"github.com/grumpycatyo-collab/go2web/business/cli"
	"github.com/grumpycatyo-collab/go2web/business/http"
	"github.com/grumpycatyo-collab/go2web/business/utils"
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
		content, err := utils.StripHTMLTags(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing response: %v\n", err)
			os.Exit(1)
		}

		filePath, err := saveToFile("output.txt", content)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error saving file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Content saved to (sorry, its too much to display in the terminal):", filePath)
		os.Exit(0)
	}
}

func saveToFile(filename, content string) (string, error) {
	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return "", err
	}
	return filename, nil
}
