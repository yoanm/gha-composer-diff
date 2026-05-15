package ddgha

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"ghaction/go-gha-wrapper/ghapi"
)

var (
	ErrClientRequired   = errors.New("client is required")
	ErrNoFilesSpecified = errors.New("at least one file must be specified")
)

type FileSpec struct {
	Repo string
	Path string
	Ref  string
}

func FetchFileContents(
	ctx context.Context,
	client *ghapi.Client,
	files map[string]FileSpec,
) (map[string][]byte, error) {
	if client == nil {
		return nil, fmt.Errorf("%w", ErrClientRequired)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("%w", ErrNoFilesSpecified)
	}

	type fetchResult struct {
		label   string
		content []byte
		err     error
	}

	resultChan := make(chan fetchResult, len(files))
	fileContents := make(map[string][]byte)

	ctx, cancelContextCb := context.WithCancel(ctx)

	routineIter := MapToCallbackIterator(
		files,
		func(label string, spec FileSpec) func() fetchResult {
			return func() fetchResult {
				slog.Debug(
					"Fetching file in background...",
					"label", label, "repo", spec.Repo, "path", spec.Path, "ref", spec.Ref,
				)

				content, err := client.LoadFileContent(ctx, spec.Repo, spec.Path, spec.Ref)
				if err != nil {
					slog.Debug("Error fetching file in background. Cancelling...",
						"label", label, "repo", spec.Repo, "path", spec.Path, "error", err,
					)

					cancelContextCb() // Stop there, no need to go further
				}

				return fetchResult{label: label, content: content, err: err}
			}
		},
	)

	collectorCb := func(res fetchResult) error {
		slog.Debug("Collecting file content...", "label", res.label)

		if res.err != nil {
			slog.Debug("Error collecting file content. Cancelling...", "label", res.label, "error", res.err)

			cancelContextCb() // Stop there, no need to go further

			return fmt.Errorf("fetching %s file content: %w", res.label, res.err)
		}

		fileContents[res.label] = res.content

		return nil
	}

	if err := RunParallelRoutines(resultChan, routineIter, collectorCb); err != nil {
		return nil, err
	}

	return fileContents, nil
}
