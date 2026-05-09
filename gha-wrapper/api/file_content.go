package api

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
)

// GitHubFileResponse represents the GitHub API response for getting file contents.
// Only the necessary fields for validation and decoding are included.
type GitHubFileResponse struct {
	Type     string `json:"type"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
}

// LoadFileContent fetches a file content from a GitHub repository for a specific Path and reference.
func (client *Client) LoadFileContent(ctx context.Context, repo string, path string, ref string) ([]byte, error) {
	slog.Debug("Load file content", "path", path, "ref", ref)

	var (
		body []byte
		err  error
	)

	url := fmt.Sprintf("/repos/%s/contents/%s?ref=%s", repo, path, ref)
	if body, err = client.httpGetRequest(ctx, url); err != nil {
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

type FileSpec struct {
	Path string
	Ref  string
}

type fetchResult struct {
	label   string
	content []byte
	err     error
}

func (client *Client) LoadMultipleFileContent(
	ctx context.Context,
	repo string,
	files map[string]FileSpec,
) (map[string][]byte, error) {
	waitGroup := sync.WaitGroup{}
	expectedCount := len(files)
	resultChan := make(chan fetchResult, expectedCount)

	ctx, cancelContextCb := context.WithCancel(ctx)

	routineCount := triggerAwaitedGoRoutines[chan fetchResult, fetchResult](
		&waitGroup,
		resultChan,
		func(yield func(func() fetchResult) bool) {
			for label, spec := range files {
				callback := func() fetchResult {
					slog.Debug("Fetching file in background...", "label", label, "path", spec.Path, "ref", spec.Ref)

					content, err := client.LoadFileContent(ctx, repo, spec.Path, spec.Ref)
					if err != nil {
						slog.Debug("Error fetching file in background. Cancelling ...", "label", label, "error", err)

						cancelContextCb() // Stop there, no need to go further
					}

					return fetchResult{label: label, content: content, err: err}
				}

				if !yield(callback) {
					return
				}
			}
		},
	)

	results := make(map[string][]byte)

	collectErrorList := collectAllAwaitedGoRoutines(
		&waitGroup,
		routineCount,
		resultChan,
		func(res fetchResult) error {
			slog.Debug("Collecting file content", "label", res.label)

			if res.err != nil {
				slog.Debug("Error collecting file content", "label", res.label, "error", res.err)

				cancelContextCb() // Stop there, no need to go further

				return fmt.Errorf("fetching %s file content: %w", res.label, res.err)
			}

			results[res.label] = res.content

			return nil
		},
	)
	if len(collectErrorList) > 0 {
		return nil, collectErrorList[0]
	}

	return results, nil
}
