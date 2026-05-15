package ghapi

import (
	"net/http"
)

type Client struct {
	client  HTTPClient
	baseUrl string
	token   string
	headers map[string]string
}

// HTTPClient interface abstracts HTTP client operations for dependency injection in tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
