package sdk

import (
	"fmt"
	"os"
)

func appendToFile(filepath string, content string) error {
	//nolint:gosec // path is controlled by upper functions (even though polluting env vars is possible)
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}

	defer file.Close()

	if _, err = file.WriteString(content); err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}
