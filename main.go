package ghaction

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"ghaction/go-gha-wrapper/ghapi"
	"ghaction/go-gha-wrapper/ghasdk"

	"github.com/yoanm/go-deps-diff/contract"
)

func New(httpClient ghapi.HTTPClient) (*Action, error) {
	requiredFields := []string{
		"manager",

		"base-sha", "head-sha",
		"base-repo", "head-repo",

		"omit-unchanged",
		"with-step-summary",
		"with-annotations",

		"gh-token",
	}

	requiredInputs, err := ghasdk.GetRequiredInputs(requiredFields)
	if err != nil {
		return nil, err //nolint:wrapcheck // Will be logged as error right away, no need to wrap it
	}

	manager := Manager(requiredInputs["manager"])

	paths, errPaths := getManagerPaths(manager)
	if errPaths != nil {
		return nil, errPaths
	}

	options := actionOptions{omitUnchanged: false, withStepSummary: false, withAnnotations: false}
	options.omitUnchanged, _ = strconv.ParseBool(requiredInputs["omit-unchanged"])
	options.withStepSummary, _ = strconv.ParseBool(requiredInputs["with-step-summary"])
	options.withAnnotations, _ = strconv.ParseBool(requiredInputs["with-annotations"])

	return &Action{
		client:  ghapi.NewClient(httpClient, os.Getenv("GITHUB_API_URL"), requiredInputs["gh-token"]),
		manager: manager,
		paths:   paths,
		refs:    refInputs{head: requiredInputs["base-sha"], base: requiredInputs["head-sha"]},
		repos:   repoInputs{head: requiredInputs["head-repo"], base: requiredInputs["base-repo"]},
		options: options,
	}, nil
}

func (gha *Action) Run() error {
	var (
		diffMap contract.DiffMap
		err     error
	)

	if diffMap, err = gha.handleDiff(context.Background()); err != nil {
		return err
	}

	if len(diffMap) == 0 {
		slog.Info("No packages detected.")

		return nil
	}

	slog.Info(fmt.Sprintf("Found %d packages", len(diffMap)))

	if gha.options.withAnnotations {
		gha.printNoticeWarning(diffMap)
	}

	return gha.handleDiffSummary(diffMap)
}

func (gha *Action) handleDiff(ctx context.Context) (contract.DiffMap, error) {
	var (
		diffMap contract.DiffMap
		err     error
	)

	switch gha.manager {
	case ComposerManager:
		diffMap, err = gha.handleComposerDiff(ctx)
	default:
		return nil, UnsupportedManagerError{gha.manager}
	}

	if err != nil {
		return nil, fmt.Errorf("handling %q diff: %w", gha.manager, err)
	}

	if gha.options.omitUnchanged {
		slog.Info("Filtering out unchanged packages...")

		for pkg, chg := range diffMap {
			if chg.Operation.Name == contract.NoChangeOperation {
				slog.Debug("Filtering out package: " + pkg)
				delete(diffMap, pkg)
			}
		}
	}

	return diffMap, nil
}

func getManagerPaths(manager Manager) (pathInputs, error) {
	paths := pathInputs{lock: "", req: ""}

	switch manager {
	case ComposerManager:
		paths.req, paths.lock = "composer.json", "composer.lock"
	default:
		return paths, UnsupportedManagerError{manager}
	}

	return paths, nil
}

type UnsupportedManagerError struct {
	manager Manager
}

func (err UnsupportedManagerError) Error() string {
	return fmt.Sprintf("unsupported manager: %s", err.manager)
}
