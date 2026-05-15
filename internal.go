package ghaction

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/yoanm/go-deps-diff/contract"
)

func (gha *Action) handleDiff(ctx context.Context) (contract.DiffMap, error) {
	var (
		diffMap contract.DiffMap
		err     error
	)

	switch gha.manager {
	case ComposerManager:
		diffMap, err = gha.handleComposerDiff(ctx)
	default:
		return nil, UnsupportedManagerError{Manager: gha.manager}
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
		return paths, UnsupportedManagerError{Manager: manager}
	}

	return paths, nil
}

func getManagerName(manager Manager) string {
	if manager == ComposerManager {
		return "Composer"
	}

	return "Generic"
}
