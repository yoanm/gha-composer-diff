package sdk

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrMissingRequiredInput = errors.New("GHA input is missing")

func GetRequiredInput(name string) (string, error) {
	val := os.Getenv("INPUT_" + strings.ToUpper(name))
	if val == "" {
		return "", fmt.Errorf("%w: %s", ErrMissingRequiredInput, name)
	}

	return val, nil
}

func GetRequiredInputs(names []string) (map[string]string, error) {
	res := make(map[string]string, len(names))
	for _, name := range names {
		val, err := GetRequiredInput(name)
		if err != nil {
			return nil, err
		}

		res[name] = val
	}

	return res, nil
}

func AppendStepSummary(content string) error {
	if err := appendToFile(os.Getenv("GITHUB_STEP_SUMMARY"), content); err != nil {
		return fmt.Errorf("writing to the summary file: %w", err)
	}

	return nil
}

func AddMultilineOutput(name string, content string) error {
	value := fmt.Sprintf("%s<<__VAR_MULILINE_EOF__\n%s\n__VAR_MULILINE_EOF__", name, content)

	if err := appendToFile(os.Getenv("GITHUB_OUTPUT"), value); err != nil {
		return fmt.Errorf("writing %q to the output file: %w", name, err)
	}

	return nil
}

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
