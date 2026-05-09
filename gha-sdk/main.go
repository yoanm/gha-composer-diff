package ghasdk

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

func GetRequiredInput(name string) string {
	val := os.Getenv("INPUT_" + strings.ToUpper(name))
	if val == "" {
		log.Fatalf("GHA input %q is missing", name)
	}

	return val
}

func AppendSummary(content string) error {
	if err := appendToFile(os.Getenv("GITHUB_STEP_SUMMARY"), content); err != nil {
		slog.Debug("error while writing to the summary file", "error", err.Error())

		return errors.New("error while writing to the summary file")
	}

	return nil
}

func SetMultilineOutput(name string, content string) error {
	value := fmt.Sprintf("%s<<__VAR_MULILINE_EOF__\n%s\n__VAR_MULILINE_EOF__", name, content)

	if err := appendToFile(os.Getenv("GITHUB_OUTPUT"), value); err != nil {
		slog.Debug("error while writing to the output file", "error", err.Error())

		return errors.New("error while writing to the output file")
	}

	return nil
}

func appendToFile(filepath string, content string) error {
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
