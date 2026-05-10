package sdk

import (
	"fmt"
	"os"
)

func AppendStepSummary(content string) error {
	if err := appendToFile(os.Getenv("GITHUB_STEP_SUMMARY"), content); err != nil {
		return fmt.Errorf("writing to the summary file: %w", err)
	}

	return nil
}
