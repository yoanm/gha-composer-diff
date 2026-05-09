package ghapi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
)

// GitHubFileResponse represents the GitHub API response for getting file contents.
// Only the necessary fields for validation and decoding are included.
type GitHubFileResponse struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

// LoadFileContent fetches a file content from a GitHub repository for a specific path and reference.
func (api *Client) LoadFileContent(repo string, path string, ref string) ([]byte, error) {
	slog.Debug("Load file content", "path", path, "ref", ref)

	var (
		body []byte
		err  error
	)

	url := fmt.Sprintf("/repos/%s/contents/%s?ref=%s", repo, path, ref)
	if body, err = api.httpGetRequest(url); err != nil {
		return nil, fmt.Errorf("failed to fetch github API: %w", err)
	}

	// Parse JSON response
	var fileResp GitHubFileResponse
	if err = json.Unmarshal(body, &fileResp); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	if fileResp.Type != "file" {
		return nil, fmt.Errorf("%w, got '%s'", ErrInvalidFileType, fileResp.Type)
	}

	if fileResp.Encoding != "base64" {
		return nil, fmt.Errorf("%w, got '%s'", ErrUnsupportedEncoding, fileResp.Encoding)
	}

	// Decode base64 content
	var decoded []byte
	if decoded, err = base64.StdEncoding.DecodeString(fileResp.Content); err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	slog.Debug("Successfully loaded file content", "path", path, "ref", ref)

	return decoded, nil
}
