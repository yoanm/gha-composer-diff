package ghadepsdiff

import (
	"fmt"

	compdiff "github.com/yoanm/go-composer-diff"
	"github.com/yoanm/go-deps-diff/contract"
	"github.com/yoanm/go-deps-diff/summary"
)

func Run(cfg *Config) (string, error) {
	var (
		diffMap contract.DiffMap
		err     error
	)

	if diffMap, err = compdiff.FileDiff(cfg.Previous, cfg.Current); err != nil {
		return "", fmt.Errorf("performing diff: %w", err)
	}

	var chgSummary string
	if len(diffMap) > 0 {
		chgSummary = summary.GenerateForChanges(diffMap)
	}

	return chgSummary, nil
}
