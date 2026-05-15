package ghasdk

import (
	"os"
	"strings"
)

type MissingRequiredInputError struct {
	Name string
}

func (err MissingRequiredInputError) Error() string {
	return "GHA input is missing: " + err.Name
}

func GetInput(name string) string {
	return os.Getenv("INPUT_" + strings.ToUpper(name))
}

func GetRequiredInput(name string) (string, error) {
	val := GetInput(name)
	if val == "" {
		return "", MissingRequiredInputError{Name: name}
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
