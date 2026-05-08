package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"
)

var (
	previousReqFileFlag  string
	previousLockFileFlag string
	currentReqFileFlag   string
	currentLockFileFlag  string
	debugModeFlag        bool
)

func main() {
	var err error

	parseFlags()

	if debugModeFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var cfg *Config
	if cfg, err = parseInputs(); err != nil {
		log.Fatal(err)
	}

	var chgSummary string
	if chgSummary, err = Run(cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Println(chgSummary)
}

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

func parseFlags() {
	flag.StringVar(&previousReqFileFlag, "previous-req", "", "Previous requirement file path")
	flag.StringVar(&previousLockFileFlag, "previous-lock", "", "Previous lock file path")
	flag.StringVar(&currentReqFileFlag, "current-req", "", "Current requirement file path")
	flag.StringVar(&currentLockFileFlag, "current-lock", "", "Current lock file path")
	flag.BoolVar(&debugModeFlag, "d", false, "Enable debug logs")

	flag.Parse()
}

func parseInputs() (*Config, error) {
	switch {
	case "" == previousReqFileFlag:
		return nil, errors.New("previous requirement file is missing")
	case "" == previousLockFileFlag:
		return nil, errors.New("previous lock file is missing")
	case "" == currentReqFileFlag:
		return nil, errors.New("current requirement file is missing")
	case "" == currentLockFileFlag:
		return nil, errors.New("current lock file is missing")
	}

	return &Config{
		Previous: &compdiff.FileInput{
			Lock:        previousLockFileFlag,
			Requirement: previousReqFileFlag,
		},
		Current: &compdiff.FileInput{
			Lock:        currentLockFileFlag,
			Requirement: currentReqFileFlag,
		},
	}, nil
}
