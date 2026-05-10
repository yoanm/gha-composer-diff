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

	basewrapper "wrapper/base-wrapper"
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
	headRepo        string
	omitUnchanged   bool
	withStepSummary bool
	withAnnotations bool
	ghToken         string // Keep this field private to avoid accidental logging !!
}

func NewActionInputs(
	lockPath, reqPath string,
	prevRef, currRef string,
	headRepo string,
	omitUnchanged bool,
	withStepSummary bool,
	withAnnotations bool,
	ghToken string,
) *ActionInputs {
	return &ActionInputs{
		lockPath:        lockPath,
		reqPath:         reqPath,
		prevRef:         prevRef,
		currRef:         currRef,
		headRepo:        headRepo,
		omitUnchanged:   omitUnchanged,
		withStepSummary: withStepSummary,
		withAnnotations: withAnnotations,
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

	if diffMap, err = handleDiff(context.Background(), client, cfg); err != nil {
		return err
	}

	if cfg.inputs.omitUnchanged {
		slog.Info("Filtering out unchanged packages...")

		for pkg, chg := range diffMap {
			if chg.Operation.Name == contract.NoChangeOperation {
				slog.Debug("Filtering out package: " + pkg)
				delete(diffMap, pkg)
			}
		}
	}

	if len(diffMap) == 0 {
		slog.Info("No packages detected.")

		return nil
	}

	slog.Info(fmt.Sprintf("Found %d packages", len(diffMap)))

	if cfg.inputs.withAnnotations {
		// 7th line is usually the "content-hash" line in the lock file. Most of time it will be updated so will show up
		// on the diff. It should at least be around the content-hash line if not exactly on it.
		// (Annotations work also if attached to unchanged line, but are less noticeable on the diff UI)
		//nolint:mnd // See above
		basewrapper.PrintNoticeWarning(diffMap, cfg.inputs.lockPath, 7)
	}

	return handleDiffSummary(diffMap, cfg.inputs.withStepSummary)
}

func handleDiff(
	ctx context.Context,
	client *api.Client,
	cfg *Config,
) (contract.DiffMap, error) {
	slog.Info("Fetching previous and current file contents...")

	fileContents, err := basewrapper.FetchFileContents(ctx, client, map[string]basewrapper.FileSpec{
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

func handleDiffSummary(diffMap contract.DiffMap, asStepSummary bool) error {
	slog.Info("Generating summary for changes...")

	chgSummary := summary.GenerateForChanges(diffMap, "Composer")

	slog.Debug("Setting summary as action output...")

	if err := sdk.AddMultilineOutput("summary", chgSummary); err != nil {
		return fmt.Errorf("setting summary as action output: %w", err)
	}

	if asStepSummary {
		slog.Info("Setting summary as step summary...")

		if err := sdk.AppendStepSummary(chgSummary); err != nil {
			return fmt.Errorf("setting summary as step summary: %w", err)
		}
	}

	slog.Debug("Change summary successfully handled")

	return nil
}
