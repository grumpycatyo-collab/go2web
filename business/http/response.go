package http

type Response struct {
	Protocol   string
	StatusCode int
	Headers    map[string]string
	Body       string
}
