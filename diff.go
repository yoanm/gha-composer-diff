package action

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"

	"action/go-gha-wrapper/api"

	ddgha "action/go-deps-diff-gha"
)

func handleDiff(
	ctx context.Context,
	client *api.Client,
	cfg *Config,
) (contract.DiffMap, error) {
	slog.Info("Fetching previous and current file contents...")

	fileContents, err := ddgha.FetchFileContents(ctx, client, map[string]ddgha.FileSpec{
		"previous-req":  {Repo: cfg.env.ghRepository, Path: cfg.inputs.reqPath, Ref: cfg.inputs.prevRef},
		"previous-lock": {Repo: cfg.env.ghRepository, Path: cfg.inputs.lockPath, Ref: cfg.inputs.prevRef},

		// Fetch current file from the head repo ! (in case of a PR from a fork)
		"current-req":  {Repo: cfg.inputs.headRepo, Path: cfg.inputs.reqPath, Ref: cfg.inputs.currRef},
		"current-lock": {Repo: cfg.inputs.headRepo, Path: cfg.inputs.lockPath, Ref: cfg.inputs.currRef},
	})
	if err != nil {
		return nil, fmt.Errorf("fetching files: %w", err)
	}

	slog.Info("Generating diff...")

	previousInput := &compdiff.Input{Lock: fileContents["previous-lock"], Requirement: fileContents["previous-req"]}
	currentInput := &compdiff.Input{Lock: fileContents["current-lock"], Requirement: fileContents["current-req"]}

	var diffMap contract.DiffMap
	if diffMap, err = compdiff.Diff(previousInput, currentInput); err != nil {
		return nil, fmt.Errorf("generating diff: %w", err)
	}

	slog.Debug("Diff generated successfully")

	return diffMap, nil
}
