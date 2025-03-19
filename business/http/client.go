package http

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"encoding/json"

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

func (c *Client) cacheResponse(url string, resp *Response) {
	data, err := json.Marshal(resp)
	if err != nil {
		return
	}

	c.cache.Set(url, data)
}

func (c *Client) getCachedResponse(url string) *Response {
	data := c.cache.Get(url)
	if data == nil {
		return nil
	}

	respBytes, ok := data.([]byte)
	if !ok {
		return nil
	}

	var cachedResp Response
	err := json.Unmarshal(respBytes, &cachedResp)
	if err != nil {
		return nil
	}

	return &cachedResp
}
func (c *Client) Get(url string) (*Response, error) {
	if cachedResp := c.getCachedResponse(url); cachedResp != nil {
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
	if parsedURL.Query != "" {
		path += "?" + parsedURL.Query
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

	var conn net.Conn
	var err2 error

	if parsedURL.Scheme == "https" {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		conn, err2 = tls.Dial("tcp", addr, tlsConfig)
	} else {
		conn, err2 = net.DialTimeout("tcp", addr, c.Timeout)
	}

	if err2 != nil {
		return nil, fmt.Errorf("connection error: %w", err2)
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

	c.cacheResponse(url, resp)

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
		Body:       "",
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
			if err == io.EOF {
				return nil, errors.New("unexpected EOF while reading headers")
			}
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
			resp.Headers[strings.ToLower(key)] = value
		}
	}

	bodyBytes, err := io.ReadAll(reader)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("error reading body: %w", err)
	}
	resp.Body = string(bodyBytes)

	return resp, nil
}
