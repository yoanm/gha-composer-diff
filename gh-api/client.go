package ghapi

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

func NewClient(
	httpClient HTTPClient,
	baseUrl string,
	token string,
) *Client {
	headers := map[string]string{
		"Accept":               "application/vnd.github+json",
		"X-Github-Api-Version": "2026-03-10",
	}

	return &Client{
		client:  httpClient,
		token:   token,
		headers: headers,
		baseUrl: baseUrl,
	}
}

// fetch performs an HTTP request to the GitHub API and returns the response body as bytes.
func (api *Client) httpGetRequest(path string) ([]byte, error) {
	slog.Debug("Fetching from GitHub API", "path", path)

	var (
		req *http.Request
		err error
	)

	// Create HTTP request
	if req, err = http.NewRequestWithContext(context.Background(), http.MethodGet, api.baseUrl+path, nil); err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add GitHub token for authentication
	if api.token != "" {
		req.Header.Set("Authorization", "Bearer "+api.token)
	}

	for key, value := range api.headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	var resp *http.Response
	if resp, err = api.client.Do(req); err != nil {
		return nil, fmt.Errorf("failed to fetch file from GitHub API: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("%w: %d %s", ErrUnexpectedStatus, resp.StatusCode, string(body))
	}

	// Read response body
	var body []byte
	if body, err = io.ReadAll(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read HTTP response: %w", err)
	}

	return body, nil
}
