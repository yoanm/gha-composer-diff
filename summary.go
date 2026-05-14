package wrapper

import (
	"fmt"
	"log/slog"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	"wrapper/go-gha-wrapper/sdk"
)

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
