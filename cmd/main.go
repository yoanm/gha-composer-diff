package main

import (
	"log/slog"
	"net/http"
	"os"

	"action/go-gha-wrapper/sdk"

	"action"
)

func main() {
	sdk.OverrideSlogDefaultLogger()

	var (
		cfg *action.Config
		err error
	)

	if cfg, err = parseInputs(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if err = action.Run(http.DefaultClient, cfg); err != nil {
		slog.Error(err.Error())
		os.Exit(2) //nolint:mnd // exit code 2 for execution error
	}
}

func parseInputs() (*action.Config, error) {
	inputs, err := sdk.GetRequiredInputs([]string{
		"lock-path",
		"req-path",
		"previous-ref",
		"current-ref",
		"head-repo",
		"omit-unchanged",
		"with-step-summary",
		"with-annotations",
		"gh-token",
	})
	if err != nil {
		return nil, err //nolint:wrapcheck // Will be logged as error right away, no need to wrap it
	}

	return action.NewConfig(
		action.NewInputsConfig(
			inputs["lock-path"]+"PLOP",
			inputs["req-path"],
			inputs["previous-ref"],
			inputs["current-ref"],
			inputs["head-repo"],
			//nolint:goconst // Mostly depend on the variable default behavior
			inputs["omit-unchanged"] != "false",
			inputs["with-step-summary"] != "false",
			inputs["with-annotations"] != "false",
			inputs["gh-token"],
		),
		action.NewEnvConfig(
			os.Getenv("GITHUB_API_URL"),
			os.Getenv("GITHUB_REPOSITORY"),
		),
	), nil
}
