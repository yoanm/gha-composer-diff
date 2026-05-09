package wrapper

import (
	"context"
	"fmt"
	"log/slog"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"

	"wrapper/gha-wrapper/api"
	"wrapper/gha-wrapper/sdk"
)

type Config struct {
	inputs *ActionInputs
	env    *ActionEnv
}

func NewConfig(inputs *ActionInputs, env *ActionEnv) *Config {
	return &Config{
		inputs: inputs,
		env:    env,
	}
}

type ActionInputs struct {
	lockPath        string
	reqPath         string
	prevRef         string
	currRef         string
	withStepSummary bool
	ghToken         string // Keep this field private to avoid accidental logging !!
}

func NewActionInputs(lockPath, reqPath, prevRef, currRef string, withStepSummary bool, ghToken string) *ActionInputs {
	return &ActionInputs{
		lockPath:        lockPath,
		reqPath:         reqPath,
		prevRef:         prevRef,
		currRef:         currRef,
		withStepSummary: withStepSummary,
		ghToken:         ghToken,
	}
}

type ActionEnv struct {
	ghAPIUrl     string
	ghRepository string
}

func NewActionEnv(ghAPIUrl, ghRepository string) *ActionEnv {
	return &ActionEnv{
		ghAPIUrl:     ghAPIUrl,
		ghRepository: ghRepository,
	}
}

func Run(httpClient api.HTTPClient, cfg *Config) error {
	client := api.NewClient(httpClient, cfg.env.ghAPIUrl, cfg.inputs.ghToken)

	var (
		diffMap contract.DiffMap
		err     error
	)
	if diffMap, err = handleDiff(client, cfg); err != nil {
		return err
	}

	if len(diffMap) == 0 {
		slog.Info("No change found")

		return nil
	}

	slog.Info(fmt.Sprintf("Found %d changes", len(diffMap)))

	return handleDiffSummary(cfg, diffMap)
}

func handleDiff(client *api.Client, cfg *Config) (contract.DiffMap, error) {
	slog.Info("Fetching previous and current file contents...")

	var (
		fileContents map[string][]byte
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

func handleDiffSummary(cfg *Config, diffMap contract.DiffMap) error {
	slog.Info("Generating summary for changes...")

	chgSummary := "# 🔎 Composer packages 🔍 \n\n" + summary.GenerateForChanges(diffMap)

	slog.Debug("Setting summary as action output...")

	if err := sdk.AddMultilineOutput("summary", chgSummary); err != nil {
		return fmt.Errorf("setting summary as action output: %w", err)
	}

	if cfg.inputs.withStepSummary {
		slog.Info("Setting summary as step summary...")

		if err := sdk.AppendStepSummary(chgSummary); err != nil {
			return fmt.Errorf("setting summary as step summary: %w", err)
		}
	}

	slog.Debug("Change summary successfully handled")

	return nil
}
