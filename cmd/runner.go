package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"

	"ghacomposerdiff/gha-wrapper/api"
	"ghacomposerdiff/gha-wrapper/sdk"
)

type fetchResult struct {
	fileType string
	content  []byte
	err      error
}

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
	)

	const numFiles = 4

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan fetchResult, numFiles)
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(numFiles)

	fetchFile := func(fileType string, path string, ref string) {
		defer waitGroup.Done()

		select {
		case <-ctx.Done():
			return
		default:
		}

		slog.Debug("fetching file", "type", fileType, "path", path, "ref", ref)
		content, err := client.LoadFileContent(repo, path, ref)

		resultChan <- fetchResult{
			fileType: fileType,
			content:  content,
			err:      err,
		}
	}

	go fetchFile("previousReq", reqRepoPath, prevRef)
	go fetchFile("previousLock", lockRepoPath, prevRef)
	go fetchFile("currentReq", reqRepoPath, currRef)
	go fetchFile("currentLock", lockRepoPath, currRef)

	resultCount := 0
	for result := range resultChan {
		resultCount++

		if result.err != nil {
			cancel()

			switch result.fileType {
			case "previousReq":
				return fmt.Errorf("fetching previous requirement file content: %w", result.err)
			case "previousLock":
				return fmt.Errorf("fetching previous lock file content: %w", result.err)
			case "currentReq":
				return fmt.Errorf("fetching current requirement file content: %w", result.err)
			case "currentLock":
				return fmt.Errorf("fetching current lock file content: %w", result.err)
			}
		}

		switch result.fileType {
		case "previousReq":
			previousReqContent = result.content
		case "previousLock":
			previousLockContent = result.content
		case "currentReq":
			currentReqContent = result.content
		case "currentLock":
			currentLockContent = result.content
		}

		if resultCount == numFiles {
			close(resultChan)

			break
		}
	}

	waitGroup.Wait()

	prevCfg := &compdiff.Input{Lock: previousLockContent, Requirement: previousReqContent}
	currCfg := &compdiff.Input{Lock: currentLockContent, Requirement: currentReqContent}

	slog.Info("Generating diff")

	var (
		diffMap contract.DiffMap
		err     error
	)

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
