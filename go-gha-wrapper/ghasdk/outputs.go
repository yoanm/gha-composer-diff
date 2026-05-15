package ghasdk

import (
	"fmt"
	"os"
)

func AddOutput(name string, content string) error {
	return appendToOutputFile(name + "=" + content)
}

func AddMultilineOutput(name string, content string) error {
	return appendToOutputFile(name + "<<" + multilineValueDelimiter + "\n" + content + "\n" + multilineValueDelimiter)
}

func appendToOutputFile(value string) error {
	if err := appendToFile(os.Getenv("GITHUB_OUTPUT"), value+"\n"); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	return nil
}
