package wrapper

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yoanm/go-deps-diff/contract"

	"wrapper/go-gha-wrapper/api"
)

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
		printNoticeWarning(diffMap, cfg.inputs.lockPath, 7)
	}

	return handleDiffSummary(diffMap, cfg.inputs.withStepSummary)
}
