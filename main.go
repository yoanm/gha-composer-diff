package action

import (
	"context"

	"github.com/yoanm/go-deps-diff/contract"

	"action/go-gha-wrapper/api"
)

func Run(httpClient api.HTTPClient, cfg *Config) error {
	client := api.NewClient(httpClient, cfg.env.ghAPIUrl, cfg.inputs.ghToken)

	var (
		diffMap contract.DiffMap
		err     error
	)

	if diffMap, err = handleDiff(context.Background(), client, cfg); err != nil {
		return err
	}

	panic("arghh")

	return handleDiffSummary(diffMap, cfg.inputs.withStepSummary)
}
