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
