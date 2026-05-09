package main

import (
	"fmt"
	"log/slog"
	"net/http"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"

	"ghacomposerdiff/gha-wrapper/api"
	"ghacomposerdiff/gha-wrapper/sdk"
)

func run(cfg *config) error {
	slog.Info("Fetch previous and current file contents")

	client := api.NewClient(
		http.DefaultClient,
		cfg.env.ghAPIUrl,
		cfg.inputs.ghToken,
	)

	reqRepoPath, lockRepoPath := cfg.inputs.reqPath, cfg.inputs.lockPath
	prevRef, currRef := cfg.inputs.prevRef, cfg.inputs.currRef
	repo := cfg.env.ghRepository

	var (
		previousReqContent  []byte
		previousLockContent []byte
		currentReqContent   []byte
		currentLockContent  []byte
		err                 error
	)

	slog.Debug("fetching previous requirement file")

	previousReqContent, err = client.LoadFileContent(repo, reqRepoPath, prevRef)
	if err != nil {
		return fmt.Errorf("fetching previous requirement file content: %w", err)
	}

	slog.Debug("fetching previous lock file")

	previousLockContent, err = client.LoadFileContent(repo, lockRepoPath, prevRef)
	if err != nil {
		return fmt.Errorf("fetching previous lock file content: %w", err)
	}

	slog.Debug("fetching current requirement file")

	currentReqContent, err = client.LoadFileContent(repo, reqRepoPath, currRef)
	if err != nil {
		return fmt.Errorf("fetching current requirement file content: %w", err)
	}

	slog.Debug("fetching current lock file")

	currentLockContent, err = client.LoadFileContent(repo, lockRepoPath, currRef)
	if err != nil {
		return fmt.Errorf("fetching current lock file content: %w", err)
	}

	prevCfg := &compdiff.Input{Lock: previousLockContent, Requirement: previousReqContent}
	currCfg := &compdiff.Input{Lock: currentLockContent, Requirement: currentReqContent}

	slog.Info("Generating diff")

	var diffMap contract.DiffMap
	if diffMap, err = compdiff.Diff(prevCfg, currCfg); err != nil {
		return fmt.Errorf("performing diff: %w", err)
	}

	slog.Debug(fmt.Sprintf("diff generated. %d changes found", len(diffMap)))

	if len(diffMap) > 0 {
		slog.Info("Generating summary for changes")

		chgSummary := "# 🔎 Composer packages 🔍 \n\n" + summary.GenerateForChanges(diffMap)

		if err2 := sdk.SetMultilineOutput("summary", chgSummary); err2 != nil {
			return fmt.Errorf("configuring action output \"summary\": %w", err2)
		}

		if cfg.inputs.withStepSummary {
			if err2 := sdk.AppendSummary(chgSummary); err2 != nil {
				return fmt.Errorf("appending change summary to the step summary: %w", err2)
			}
		}
	}

	return nil
}
