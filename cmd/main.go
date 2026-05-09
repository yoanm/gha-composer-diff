package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	ghapi "ghacomposerdiff/gh-api"
	ghasdk "ghacomposerdiff/gha-sdk"

	summary "github.com/yoanm/go-deps-diff-summary"
	"github.com/yoanm/go-deps-diff/contract"

	compdiff "github.com/yoanm/go-composer-diff"
)

type config struct {
	inputs actionInputs
	env    actionEnv
}
type actionInputs struct {
	lockPath        string
	reqPath         string
	prevRef         string
	currRef         string
	withStepSummary bool
}
type actionEnv struct {
	ghRepository string
	ghAPIUrl     string
	ghToken      string
}

func main() {
	if os.Getenv("RUNNER_DEBUG") == "1" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	var (
		cfg *config
		err error
	)

	if cfg, err = parseInputs(); err != nil {
		log.Fatal(err)
	}

	if err = Run(cfg); err != nil {
		log.Fatal(err)
	}
}

func Run(cfg *config) error {
	slog.Info("Fetch previous and current file contents")

	client := ghapi.NewClient(http.DefaultClient, cfg.env.ghAPIUrl, cfg.env.ghToken)

	reqRepoPath, lockRepoPath := cfg.inputs.reqPath, cfg.inputs.lockPath
	prevRef, currRef := cfg.inputs.prevRef, cfg.inputs.currRef
	repo := cfg.env.ghRepository

	var (
		previousReqContent  []byte
		previousLockContent []byte
		currentReqContent   []byte
		currentLockContent  []byte
		err                 error
	)

	slog.Debug("fetching previous requirement file")

	previousReqContent, err = client.LoadFileContent(repo, reqRepoPath, prevRef)
	if err != nil {
		return fmt.Errorf("fetching previous requirement file content: %w", err)
	}

	slog.Debug("fetching previous lock file")

	previousLockContent, err = client.LoadFileContent(repo, lockRepoPath, prevRef)
	if err != nil {
		return fmt.Errorf("fetching previous lock file content: %w", err)
	}

	slog.Debug("fetching current requirement file")

	currentReqContent, err = client.LoadFileContent(repo, reqRepoPath, currRef)
	if err != nil {
		return fmt.Errorf("fetching current requirement file content: %w", err)
	}

	slog.Debug("fetching current lock file")

	currentLockContent, err = client.LoadFileContent(repo, lockRepoPath, currRef)
	if err != nil {
		return fmt.Errorf("fetching current lock file content: %w", err)
	}

	prevCfg := &compdiff.Input{Lock: previousLockContent, Requirement: previousReqContent}
	currCfg := &compdiff.Input{Lock: currentLockContent, Requirement: currentReqContent}

	slog.Info("Generating diff")

	var diffMap contract.DiffMap
	if diffMap, err = compdiff.Diff(prevCfg, currCfg); err != nil {
		return fmt.Errorf("performing diff: %w", err)
	}

	slog.Debug(fmt.Sprintf("diff generated. %d changes found", len(diffMap)))

	if len(diffMap) > 0 {
		slog.Info("Generating summary for changes")

		chgSummary := "# 🔎 Composer packages 🔍 \n\n" + summary.GenerateForChanges(diffMap)

		if err2 := ghasdk.SetMultilineOutput("summary", chgSummary); err2 != nil {
			return fmt.Errorf("configuring action output \"summary\": %w", err2)
		}

		if cfg.inputs.withStepSummary {
			if err2 := ghasdk.AppendSummary(chgSummary); err2 != nil {
				return fmt.Errorf("appending change summary to the step summary: %w", err2)
			}
		}
	}

	return nil
}

func parseInputs() (*config, error) {
	return &config{
		inputs: actionInputs{
			lockPath:        ghasdk.GetRequiredInput("lock-path"),
			reqPath:         ghasdk.GetRequiredInput("req-path"),
			prevRef:         ghasdk.GetRequiredInput("previous-ref"),
			currRef:         ghasdk.GetRequiredInput("current-ref"),
			withStepSummary: ghasdk.GetRequiredInput("with-step-summary") == "true",
		},
		env: actionEnv{
			ghRepository: os.Getenv("GITHUB_REPOSITORY"),
			ghAPIUrl:     os.Getenv("GITHUB_API_URL"),
			ghToken:      os.Getenv("GITHUB_TOKEN"),
		},
	}, nil
}
