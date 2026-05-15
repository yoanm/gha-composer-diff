package main

import (
	"log/slog"
	"net/http"
	"os"

	"ghaction"
	"ghaction/go-gha-wrapper/ghasdk"
)

func main() {
	ghasdk.OverrideSlogDefaultLogger()

	action, err := ghaction.New(http.DefaultClient)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	if err = action.Run(); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
