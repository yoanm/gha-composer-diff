package basewrapper

import (
	"context"
	"fmt"
	"iter"
	"log/slog"

	"wrapper/gha-wrapper/api"
)

type FileSpec struct {
	Repo string
	Path string
	Ref  string
}

func FetchFileContents(ctx context.Context, client *api.Client, files map[string]FileSpec) (map[string][]byte, error) {
	type fetchResult struct {
		label   string
		content []byte
		err     error
	}

	resultChan := make(chan fetchResult, len(files))
	fileContents := make(map[string][]byte)

	ctx, cancelContextCb := context.WithCancel(ctx)

	var routineIter iter.Seq[func() fetchResult] = func(yield func(func() fetchResult) bool) {
		for label, spec := range files {
			callback := func() fetchResult {
				slog.Debug("Fetching file in background...", "label", label, "path", spec.Path, "ref", spec.Ref)

				content, err := client.LoadFileContent(ctx, spec.Repo, spec.Path, spec.Ref)
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
	}

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
		return nil, fmt.Errorf("fetching files: %w", err)
	}

	return fileContents, nil
}
