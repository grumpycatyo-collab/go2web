package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Options struct {
	URL        string
	SearchTerm string
	Help       bool
}

func ParseArgs(args []string) (Options, error) {
	options := Options{}

	if len(args) == 0 {
		return options, errors.New("no arguments provided")
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-u":
			if i+1 < len(args) {
				options.URL = args[i+1]
				i++
			} else {
				return options, errors.New("URL not provided after -u flag")
			}
		case "-s":
			if i+1 < len(args) {
				options.SearchTerm = strings.Join(args[i+1:], " ")
				i = len(args)
			} else {
				return options, errors.New("search term not provided after -s flag")
			}
		case "--help":
			options.Help = true
		default:
			return options, fmt.Errorf("unknown option: %s", args[i])
		}
	}

	return options, nil
}

func PrintHelp() {
	fmt.Fprintf(os.Stderr, `Usage: go2web [OPTIONS]

Options:
  -u <URL>         Make an HTTP request to the specified URL and print the response
  -s <search-term> Make an HTTP request to search the term using a search engine and print top 10 results
  --help               Show this help message

Examples:
  go2web -u https://example.com
  go2web -s golang websockets
`)
}
