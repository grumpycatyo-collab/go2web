package http

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/grumpycatyo-collab/go2web/business/cache"
	"github.com/grumpycatyo-collab/go2web/business/utils"
)

const (
	DefaultUserAgent = "go2web/1.0"
	DefaultTimeout   = 30 * time.Second
)

type Client struct {
	UserAgent  string
	Timeout    time.Duration
	cache      *cache.Cache
	maxRetries int
}

func NewClient() *Client {
	return &Client{
		UserAgent:  DefaultUserAgent,
		Timeout:    DefaultTimeout,
		cache:      cache.NewCache(),
		maxRetries: 5,
	}
}

func (c *Client) Get(url string) (*Response, error) {
	if cachedResp, ok := c.cache.Get(url).(*Response); ok {
		return cachedResp, nil
	}

	parsedURL, err := utils.ParseURL(url)
	if err != nil {
		return nil, err
	}
	host := parsedURL.Host
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}
	port := "80"
	if parsedURL.Scheme == "https" {
		port = "443"
	}
	if strings.Contains(host, ":") {
		parts := strings.Split(host, ":")
		host = parts[0]
		port = parts[1]
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	conn, err := net.DialTimeout("tcp", addr, c.Timeout)
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	defer conn.Close()

	req := &Request{
		Method:  "GET",
		Path:    path,
		Host:    host,
		Headers: map[string]string{},
	}
	req.Headers["Host"] = host
	req.Headers["User-Agent"] = c.UserAgent
	req.Headers["Connection"] = "close"
	req.Headers["Accept"] = "text/html, application/json"

	reqStr := req.String()
	_, err = conn.Write([]byte(reqStr))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	resp, err := readResponse(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}
	fmt.Println(resp)

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Headers["Location"]
		if location == "" {
			return nil, errors.New("redirect response without Location header")
		}

		if !strings.HasPrefix(location, "http") {
			if strings.HasPrefix(location, "/") {
				location = fmt.Sprintf("%s://%s%s", parsedURL.Scheme, host, location)
			} else {
				location = fmt.Sprintf("%s://%s/%s", parsedURL.Scheme, host, location)
			}
		}

		return c.Get(location)
	}

	c.cache.Set(url, resp)

	return resp, nil
}

func readResponse(conn net.Conn) (*Response, error) {
	reader := bufio.NewReader(conn)

	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("error reading status line: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(statusLine), " ")
	if len(parts) < 2 {
		return nil, errors.New("invalid status line")
	}

	resp := &Response{
		Protocol:   parts[0],
		StatusCode: 0,
		Headers:    map[string]string{},
	}

	if len(parts) > 1 {
		statusCode, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid status code: %w", err)
		}
		resp.StatusCode = statusCode
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error reading header: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			break
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx > 0 {
			key := strings.TrimSpace(line[:colonIdx])
			value := strings.TrimSpace(line[colonIdx+1:])
			resp.Headers[key] = value
		}
	}

	if chunked := resp.Headers["Transfer-Encoding"] == "chunked"; chunked {
		var body strings.Builder
		for {
			sizeLine, err := reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading chunk size: %w", err)
			}

			size, err := strconv.ParseInt(strings.TrimSpace(sizeLine), 16, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid chunk size: %w", err)
			}

			if size == 0 {
				_, err = reader.ReadString('\n')
				if err != nil {
					return nil, fmt.Errorf("error reading final CRLF: %w", err)
				}
				break
			}

			chunk := make([]byte, size)
			_, err = reader.Read(chunk)
			if err != nil {
				return nil, fmt.Errorf("error reading chunk data: %w", err)
			}

			body.Write(chunk)

			_, err = reader.ReadString('\n')
			if err != nil {
				return nil, fmt.Errorf("error reading chunk CRLF: %w", err)
			}
		}

		resp.Body = body.String()
	} else if contentLength, ok := resp.Headers["Content-Length"]; ok {
		length, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length: %w", err)
		}

		bodyBytes := make([]byte, length)
		_, err = reader.Read(bodyBytes)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}

		resp.Body = string(bodyBytes)
	} else {
		bodyBytes, err := reader.ReadBytes(0)
		if err != nil && err.Error() != "EOF" {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
		resp.Body = string(bodyBytes)
	}

	return resp, nil
}
