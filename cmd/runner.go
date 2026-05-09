package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"

	"ghacomposerdiff/gha-wrapper/api"
	"ghacomposerdiff/gha-wrapper/sdk"
)

type config struct {
	inputs *actionInputs
	env    *actionEnv
}

type actionInputs struct {
	lockPath        string
	reqPath         string
	prevRef         string
	currRef         string
	withStepSummary bool
	ghToken         string // Keep this field as private to avoid accidental logging !!
}

type actionEnv struct {
	ghRepository string
	ghAPIUrl     string
}

func run(cfg *config) error {
	slog.Info("Fetching previous and current file contents...")

	client := api.NewClient(http.DefaultClient, cfg.env.ghAPIUrl, cfg.inputs.ghToken)

	var (
		fileContents map[string][]byte
		diffMap      contract.DiffMap
		err          error
	)

	fileContents, err = client.LoadMultipleFileContent(
		context.Background(),
		cfg.env.ghRepository,
		map[string]api.FileSpec{
			"previous-req":  {Path: cfg.inputs.reqPath, Ref: cfg.inputs.prevRef},
			"previous-lock": {Path: cfg.inputs.lockPath, Ref: cfg.inputs.prevRef},
			"current-req":   {Path: cfg.inputs.reqPath, Ref: cfg.inputs.currRef},
			"current-lock":  {Path: cfg.inputs.lockPath, Ref: cfg.inputs.currRef},
		},
	)
	if err != nil {
		return fmt.Errorf("loading files: %w", err)
	}

	slog.Info("Generating diff...")

	diffMap, err = compdiff.Diff(
		&compdiff.Input{Lock: fileContents["previous-lock"], Requirement: fileContents["previous-req"]},
		&compdiff.Input{Lock: fileContents["current-lock"], Requirement: fileContents["current-req"]},
	)
	if err != nil {
		return fmt.Errorf("performing diff: %w", err)
	}

	if len(diffMap) == 0 {
		slog.Info("No change found")

		return nil
	}

	slog.Info(fmt.Sprintf("Found %d changes", len(diffMap)))
	slog.Info("Generating summary for changes")

	chgSummary := "# 🔎 Composer packages 🔍 \n\n" + summary.GenerateForChanges(diffMap)

	if err = sdk.SetMultilineOutput("summary", chgSummary); err != nil {
		return fmt.Errorf("configuring action output \"summary\": %w", err)
	}

	if cfg.inputs.withStepSummary {
		slog.Info("Appending summary as step summary")

		if err = sdk.AppendSummary(chgSummary); err != nil {
			return fmt.Errorf("appending change summary to the step summary: %w", err)
		}
	}

	return nil
}
