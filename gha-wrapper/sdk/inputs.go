package sdk

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrMissingRequiredInput = errors.New("GHA input is missing")

func GetInput(name string) string {
	return os.Getenv("INPUT_" + strings.ToUpper(name))
}

func GetRequiredInput(name string) (string, error) {
	val := GetInput(name)
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
