package http

import (
	"fmt"
	"strings"
)

type Request struct {
	Method  string
	Path    string
	Host    string
	Headers map[string]string
	Body    string
}

func (r *Request) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s %s HTTP/1.1\r\n", r.Method, r.Path))

	for key, value := range r.Headers {
		sb.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}

	sb.WriteString("\r\n")

	if r.Body != "" {
		sb.WriteString(r.Body)
	}

	return sb.String()
}
