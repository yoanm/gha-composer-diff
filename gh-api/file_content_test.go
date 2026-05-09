package ghapi_test

import (
	"encoding/base64"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"

	ghapi "ghacomposerdiff/gh-api"
)

// mockHTTPClient is a mock implementation of HTTPClient for testing.
type mockHTTPClient struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

// compile-time assertion that mockHTTPClient implements HTTPClient.
var _ ghapi.HTTPClient = (*mockHTTPClient)(nil)

// newMockResponse creates a minimal http.Response with all required fields.
func newMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{ //nolint:exhaustruct // Useless fields are omitted for brevity.
		Status:        http.StatusText(statusCode),
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(strings.NewReader(body)),
		ContentLength: -1,
		Header:        make(http.Header),
	}
}

// TestLoadFileContent_Success tests successful file fetching with valid type and encoding.
func TestLoadFileContent_Success(t *testing.T) {
	t.Parallel()

	expectedContent := []byte("hello world\nthis is a test file")
	encodedContent := base64.StdEncoding.EncodeToString(expectedContent)

	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			body := `{"type":"file","encoding":"base64","content":"` + encodedContent + `"}`

			return newMockResponse(http.StatusOK, body), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

	content, err := client.LoadFileContent("owner/repo", "file.txt", "main")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	} else if !slices.Equal(content, expectedContent) {
		t.Errorf("expected %q, got %q", expectedContent, content)
	}
}

// TestLoadFileContent_TypeNotFile tests error when type is not "file".
func TestLoadFileContent_TypeNotFile(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			enc := base64.StdEncoding.EncodeToString([]byte("content"))
			body := `{"type":"dir","encoding":"base64","content":"` + enc + `"}`

			return newMockResponse(http.StatusOK, body), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

	_, err := client.LoadFileContent("owner/repo", "path", "main")
	if err == nil || !strings.Contains(err.Error(), "invalid response type") {
		t.Errorf("expected type validation error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "expected 'file'") {
		t.Errorf("expected 'expected file' in error message, got: %v", err)
	}
}

// TestLoadFileContent_EncodingNotBase64 tests error when encoding is not "base64".
func TestLoadFileContent_EncodingNotBase64(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			body := `{"type":"file","encoding":"utf-8","content":"raw content"}`

			return newMockResponse(http.StatusOK, body), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

	_, err := client.LoadFileContent("owner/repo", "file.txt", "main")
	if err == nil || !strings.Contains(err.Error(), "unsupported encoding") {
		t.Errorf("expected encoding validation error, got: %v", err)
	}

	if !strings.Contains(err.Error(), "expected 'base64'") {
		t.Errorf("expected 'expected base64' in error message, got: %v", err)
	}
}

// TestLoadFileContent_HTTP404 tests handling of 404 Not Found.
func TestLoadFileContent_HTTP404(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return newMockResponse(http.StatusNotFound, `{"message":"Not Found"}`), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

	_, err := client.LoadFileContent("owner/repo", "file.txt", "main")
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("expected 404 error, got: %v", err)
	}
}

// TestLoadFileContent_HTTP401 tests handling of 401 Unauthorized.
func TestLoadFileContent_HTTP401(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			return newMockResponse(http.StatusUnauthorized, `{"message":"Bad credentials"}`), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

	_, err := client.LoadFileContent("owner/repo", "file.txt", "main")
	if err == nil || !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 error, got: %v", err)
	}
}

// TestLoadFileContent_InvalidBase64 tests handling of invalid base64 content.
func TestLoadFileContent_InvalidBase64(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			body := `{"type":"file","encoding":"base64","content":"not-valid-base64!!!"}`

			return newMockResponse(http.StatusOK, body), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

	_, err := client.LoadFileContent("owner/repo", "file.txt", "main")
	if err == nil || !strings.Contains(err.Error(), "failed to decode base64") {
		t.Errorf("expected base64 error, got: %v", err)
	}
}

// TestLoadFileContent_URLConstruction tests URL construction.
func TestLoadFileContent_URLConstruction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		repo        string
		path        string
		ref         string
		expectedURL string
	}{
		{
			"basic", "owner/repo", "file.txt", "main",
			"https://api.github.com/repos/owner/repo/contents/file.txt?ref=main",
		},
		{
			"nested", "org/repo", "src/file.go", "develop",
			"https://api.github.com/repos/org/repo/contents/src/file.go?ref=develop",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var url string

			mock := &mockHTTPClient{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					url = req.URL.String()
					enc := base64.StdEncoding.EncodeToString([]byte("test"))
					body := `{"type":"file","encoding":"base64","content":"` + enc + `"}`

					return newMockResponse(http.StatusOK, body), nil
				},
			}

			client := ghapi.NewClient(mock, "https://api.github.com", "test-token")

			_, err := client.LoadFileContent(testCase.repo, testCase.path, testCase.ref)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if url != testCase.expectedURL {
				t.Errorf("expected url %q, got %q", testCase.expectedURL, url)
			}
		})
	}
}

// TestLoadFileContent_HeaderVerification tests request headers.
func TestLoadFileContent_HeaderVerification(t *testing.T) {
	t.Parallel()

	headers := make(map[string]string)
	mock := &mockHTTPClient{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			headers["Accept"] = req.Header.Get("Accept")
			headers["X-Github-Api-Version"] = req.Header.Get("X-Github-Api-Version")
			headers["Authorization"] = req.Header.Get("Authorization")

			enc := base64.StdEncoding.EncodeToString([]byte("test"))
			body := `{"type":"file","encoding":"base64","content":"` + enc + `"}`

			return newMockResponse(http.StatusOK, body), nil
		},
	}

	client := ghapi.NewClient(mock, "https://api.github.com", "a-token")

	_, err := client.LoadFileContent("owner/repo", "file.txt", "main")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if headers["Accept"] != "application/vnd.github+json" {
		t.Errorf("expected Accept header, got %q", headers["Accept"])
	}

	if headers["X-Github-Api-Version"] != "2026-03-10" {
		t.Errorf("expected API version, got %q", headers["X-Github-Api-Version"])
	}

	if headers["Authorization"] != "Bearer a-token" {
		t.Errorf("expected auth header, got %q", headers["Authorization"])
	}
}

// TestLoadFileContent_VariousContentTypes tests various content.
func TestLoadFileContent_VariousContentTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{"simple", "hello"},
		{"multiline", "line1\nline2"},
		{"unicode", "Hello world"},
		{"empty", ""},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			enc := base64.StdEncoding.EncodeToString([]byte(testCase.content))
			mock := &mockHTTPClient{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					body := `{"type":"file","encoding":"base64","content":"` + enc + `"}`

					return newMockResponse(http.StatusOK, body), nil
				},
			}

			client := ghapi.NewClient(mock, "https://api.github.com", "a-token")

			content, err := client.LoadFileContent("owner/repo", "file.txt", "main")
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if string(content) != testCase.content {
				t.Errorf("expected %q, got %q", testCase.content, string(content))
			}
		})
	}
}
