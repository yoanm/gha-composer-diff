package ghaction

import (
	"context"
	"fmt"
	"log/slog"

	ddgha "ghaction/go-deps-diff-gha"

	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"
)

func (gha *Action) handleComposerDiff(ctx context.Context) (contract.DiffMap, error) {
	slog.Info("Composer: Fetching previous and current file contents...")

	fileContents, err := ddgha.FetchFileContents(ctx, gha.client, map[string]ddgha.FileSpec{
		"previous-req":  {Repo: gha.repos.base, Path: gha.paths.req, Ref: gha.refs.base},
		"previous-lock": {Repo: gha.repos.base, Path: gha.paths.lock, Ref: gha.refs.base},
		// ! Fetch current file from the head repo ! (in case of a PR from a fork)
		"current-req":  {Repo: gha.repos.head, Path: gha.paths.req, Ref: gha.refs.head},
		"current-lock": {Repo: gha.repos.head, Path: gha.paths.lock, Ref: gha.refs.head},
	})
	if err != nil {
		return nil, fmt.Errorf("fetching files: %w", err)
	}

	slog.Info("Composer: Generating diff...")

	previousInput := &compdiff.Input{Lock: fileContents["previous-lock"], Requirement: fileContents["previous-req"]}
	currentInput := &compdiff.Input{Lock: fileContents["current-lock"], Requirement: fileContents["current-req"]}

	var diffMap contract.DiffMap
	if diffMap, err = compdiff.Diff(previousInput, currentInput); err != nil {
		return nil, fmt.Errorf("generating diff: %w", err)
	}

	slog.Debug("Composer: Diff generated successfully")

	return diffMap, nil
}
