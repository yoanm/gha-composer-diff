package sdk

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
)

type LogHandler struct {
	writer      io.Writer
	attrHandler slog.Handler
	mutex       *sync.Mutex
	attrBuffer  *bytes.Buffer
}

var _ slog.Handler = (*LogHandler)(nil) // Ensure LogHandler implements slog.Handler

func OverrideDefaultLogger() {
	level := slog.LevelInfo
	if os.Getenv("RUNNER_DEBUG") == "1" {
		level = slog.LevelDebug
	}

	attrBuffer := &bytes.Buffer{}
	handler := &LogHandler{
		writer: os.Stderr,
		attrHandler: slog.NewTextHandler(
			attrBuffer,
			&slog.HandlerOptions{ //nolint:exhaustruct // Only useful options are defined
				Level: level,
				ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					// Remove default attributes so only user-defined attributes are logged
					if attr.Key == slog.TimeKey || attr.Key == slog.LevelKey || attr.Key == slog.MessageKey {
						return slog.Attr{} //nolint:exhaustruct // Return empty attribute to skip logging it
					}

					return attr
				},
			},
		),
		mutex:      &sync.Mutex{},
		attrBuffer: attrBuffer,
	}

	slog.SetDefault(slog.New(handler))
	slog.SetLogLoggerLevel(slog.LevelInfo) // Set default as info (mostly when using log.Println, etc.)
}

func (handler *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.attrHandler.Enabled(ctx, level)
}

func (handler *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	handler.mutex.Lock()
	defer func() {
		handler.attrBuffer.Reset()
		handler.mutex.Unlock()
	}()

	var msg string
	if record.Level != slog.LevelDebug && record.Level != slog.LevelWarn && record.Level != slog.LevelError {
		msg = record.Message
	} else {
		var level string

		switch record.Level {
		case slog.LevelDebug:
			level = "debug"
		case slog.LevelWarn:
			level = "warning"
		case slog.LevelError:
			level = "error"
		default:
			panic("unexpected log level: " + record.Level.String())
		}

		msg = "::" + level + "::" + record.Message
	}

	// Delegate potential attributes and end of line management to the attrHandler handler
	err := handler.attrHandler.Handle(ctx, record)
	if err != nil {
		return err //nolint:wrapcheck // Current logger is just a proxy
	}

	attrString := handler.attrBuffer.String()
	if len(attrString) > 1 { // More than one char (the newline char) means there are attributes to log
		attrString = " " + attrString // Add a space before attributes if they exist
	}

	_, err = fmt.Fprint(handler.writer, msg+attrString)

	return err //nolint:wrapcheck // Current logger is just a proxy
}

func (handler *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler.attrHandler.WithAttrs(attrs)
}

func (handler *LogHandler) WithGroup(name string) slog.Handler {
	return handler.attrHandler.WithGroup(name)
}
