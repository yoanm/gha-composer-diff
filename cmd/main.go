package main

import (
	"log/slog"
	"net/http"
	"os"

	"wrapper/gha-wrapper/sdk"

	"wrapper"
)

func main() {
	sdk.OverrideDefaultLogger()

	var (
		cfg *wrapper.Config
		err error
	)

	if cfg, err = parseInputs(); err != nil {
		slog.Error(err.Error())
		os.Exit(2) //nolint:mnd // exit code 2 for input parsing error
	}

	if err = wrapper.Run(http.DefaultClient, cfg); err != nil {
		slog.Error(err.Error())
		os.Exit(3) //nolint:mnd // exit code 3 for execution error
	}
}

func parseInputs() (*wrapper.Config, error) {
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

	return wrapper.NewConfig(
		wrapper.NewActionInputs(
			inputs["lock-path"],
			inputs["req-path"],
			inputs["previous-ref"],
			inputs["current-ref"],
			inputs["with-step-summary"] == "true",
			inputs["gh-token"],
		),
		wrapper.NewActionEnv(
			os.Getenv("GITHUB_API_URL"),
			os.Getenv("GITHUB_REPOSITORY"),
		),
	), nil
}
