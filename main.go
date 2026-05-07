package ghadepsdiff

import (
	"fmt"
	"os"

	depsdiff "github.com/yoanm/go-deps-diff"
	"github.com/yoanm/go-deps-diff/shared"
	"github.com/yoanm/go-deps-diff/summary"
)

func Run(cfg *Config) (string, error) {
	var (
		inputPrevious *depsdiff.PkgManagerInput
		inputCurrent  *depsdiff.PkgManagerInput
		err           error
	)

	if inputPrevious, err = loadFiles(cfg.Previous.Requirement, cfg.Previous.Lock); err != nil {
		return "", fmt.Errorf("reading previous files: %w", err)
	}

	if inputCurrent, err = loadFiles(cfg.Current.Requirement, cfg.Current.Lock); err != nil {
		return "", fmt.Errorf("reading current files: %w", err)
	}

	var diffMap shared.DiffMap
	if diffMap, err = performDiff(cfg, inputPrevious, inputCurrent); err != nil {
		return "", fmt.Errorf("performing diff: %w", err)
	}

	var chgSummary string
	if len(diffMap) > 0 {
		chgSummary = summary.GenerateForChanges(diffMap)
	}

	return chgSummary, nil
}

func loadFiles(reqPath, lockPath string) (*depsdiff.PkgManagerInput, error) {
	previousReqContent, err := os.ReadFile(reqPath)
	if err != nil {
		return nil, fmt.Errorf("reading requirement file: %w", err)
	}

	previousLockContent, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, fmt.Errorf("reading lock file: %w", err)
	}

	return &depsdiff.PkgManagerInput{
		Lock:        previousLockContent,
		Requirement: previousReqContent,
	}, nil
}

func performDiff(cfg *Config, prev *depsdiff.PkgManagerInput, curr *depsdiff.PkgManagerInput) (shared.DiffMap, error) {
	switch cfg.Manager {
	case ComposerManager:
		return depsdiff.ComposerDiff(prev, curr)
	default:
		return nil, fmt.Errorf("unknown manager: %s", cfg.Manager)
	}
}
