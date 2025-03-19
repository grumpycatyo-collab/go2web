package utils

import (
	"errors"
	"strings"
)

type URL struct {
	Scheme string
	Host   string
	Path   string
	Query  string
}

func ParseURL(rawURL string) (*URL, error) {
	url := &URL{}

	schemeIdx := strings.Index(rawURL, "://")
	if schemeIdx == -1 {
		url.Scheme = "http"
		rawURL = "http://" + rawURL
		schemeIdx = strings.Index(rawURL, "://")
	} else {
		url.Scheme = rawURL[:schemeIdx]
	}

	hostStart := schemeIdx + 3
	pathIdx := strings.IndexByte(rawURL[hostStart:], '/')

	if pathIdx == -1 {
		url.Host = rawURL[hostStart:]
		url.Path = "/"
	} else {
		url.Host = rawURL[hostStart : hostStart+pathIdx]

		queryIdx := strings.IndexByte(rawURL[hostStart+pathIdx:], '?')
		if queryIdx == -1 {
			url.Path = rawURL[hostStart+pathIdx:]
		} else {
			url.Path = rawURL[hostStart+pathIdx : hostStart+pathIdx+queryIdx]
			url.Query = rawURL[hostStart+pathIdx+queryIdx+1:]
		}
	}

	if url.Host == "" {
		return nil, errors.New("invalid URL: missing host")
	}

	return url, nil
}
