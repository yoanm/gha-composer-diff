package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"

	compdiff "github.com/yoanm/go-composer-diff"

	"ghadepsdiff"
)

var (
	previousReqFileFlag  string
	previousLockFileFlag string
	currentReqFileFlag   string
	currentLockFileFlag  string
	debugModeFlag        bool
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var err error

	parseFlags()

	if debugModeFlag {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var cfg *ghadepsdiff.Config
	if cfg, err = parseInputs(); err != nil {
		return err
	}

	var chgSummary string
	if chgSummary, err = ghadepsdiff.Run(cfg); err != nil {
		return err
	}

	fmt.Println(chgSummary)

	return nil
}

func parseFlags() {
	flag.StringVar(&previousReqFileFlag, "previous-req", "", "Previous requirement file path")
	flag.StringVar(&previousLockFileFlag, "previous-lock", "", "Previous lock file path")
	flag.StringVar(&currentReqFileFlag, "current-req", "", "Current requirement file path")
	flag.StringVar(&currentLockFileFlag, "current-lock", "", "Current lock file path")
	flag.BoolVar(&debugModeFlag, "d", false, "Enable debug logs")

	flag.Parse()
}

func parseInputs() (*ghadepsdiff.Config, error) {
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

	return &ghadepsdiff.Config{
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
