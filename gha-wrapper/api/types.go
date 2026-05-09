package api

import (
	"errors"
	"net/http"
)

// Sentinel errors for GitHub API operations.
var (
	ErrInvalidFileType     = errors.New("invalid response type: expected 'file'")
	ErrUnsupportedEncoding = errors.New("unsupported encoding: expected 'base64'")
	ErrUnexpectedStatus    = errors.New("unexpected HTTP status code from GitHub API")
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
