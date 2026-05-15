package ghaction

import (
	"fmt"
	"log/slog"

	"ghaction/go-gha-wrapper/ghasdk"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"
)

func (gha *Action) handleDiffSummary(diffMap contract.DiffMap) error {
	slog.Info("Generating summary for changes...")

	chgSummary := summary.GenerateForChanges(diffMap, getManagerName(gha.manager))

	slog.Debug("Setting summary as action output...")

	if err := ghasdk.AddMultilineOutput("summary", chgSummary); err != nil {
		return fmt.Errorf("setting summary as action output: %w", err)
	}

	if gha.options.withStepSummary {
		slog.Info("Setting summary as step summary...")

		if err := ghasdk.AppendStepSummary(chgSummary); err != nil {
			return fmt.Errorf("setting summary as step summary: %w", err)
		}
	}

	slog.Debug("Change summary successfully handled")

	return nil
}
