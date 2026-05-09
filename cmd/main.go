package main

import (
	"log/slog"
	"os"

	"ghacomposerdiff/gha-wrapper/sdk"
)

func main() {
	sdk.OverrideDefaultLogger()

	var (
		cfg *config
		err error
	)

	if cfg, err = parseInputs(); err != nil {
		slog.Error(err.Error())
		os.Exit(2) //nolint:mnd // exit code 2 for input parsing error
	}

	if err = run(cfg); err != nil {
		slog.Error(err.Error())
		os.Exit(3) //nolint:mnd // exit code 3 for execution error
	}
}

func parseInputs() (*config, error) {
	inputs, err := sdk.GetRequiredInputs([]string{
		"lock-path",
		"req-path",
		"previous-ref",
		"current-ref",
		"with-step-summary",
		"gh-token",
	})
	if err != nil {
		return nil, err //nolint:wrapcheck // Will be logged as error right away, no need to wrap it
	}

	return &config{
		inputs: &actionInputs{
			lockPath:        inputs["lock-path"],
			reqPath:         inputs["req-path"],
			prevRef:         inputs["previous-ref"],
			currRef:         inputs["current-ref"],
			withStepSummary: inputs["with-step-summary"] == "true",
			ghToken:         inputs["gh-token"],
		},
		env: &actionEnv{
			ghRepository: os.Getenv("GITHUB_REPOSITORY"),
			ghAPIUrl:     os.Getenv("GITHUB_API_URL"),
		},
	}, nil
}
