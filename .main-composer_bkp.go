package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"
)

type Config struct {
	Previous *compdiff.FileInput
	Current  *compdiff.FileInput
}

var (
	previousReqFileFlag  string
	previousLockFileFlag string
	currentReqFileFlag   string
	currentLockFileFlag  string
	debugModeFlag        bool
)

func main() {
	parseFlags()

	if debugModeFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var (
		cfg *Config
		err error
	)

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
	lockPath, reqPath := os.Getenv("INPUT_LOCK-PATH"), os.Getenv("INPUT_REQ-PATH")
	prevRef, currRef := os.Getenv("INPUT_PREVIOUS-REF"), os.Getenv("INPUT_CURRENT-REF")
	switch {
	case "" == lockPath:
		return nil, errors.New("lock-path input is missing")
	case "" == reqPath:
		return nil, errors.New("req-path input is missing")
	case "" == prevRef:
		return nil, errors.New("previous-ref input is missing")
	case "" == currRef:
		return nil, errors.New("current-ref input is missing")
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
