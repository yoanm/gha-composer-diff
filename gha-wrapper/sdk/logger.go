package sdk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type LogHandler struct {
	writer   io.Writer
	original slog.Handler
}

var _ slog.Handler = (*LogHandler)(nil) // Ensure LogHandler implements slog.Handler

func OverrideDefaultLogger() {
	level := slog.LevelInfo
	if os.Getenv("RUNNER_DEBUG") == "1" {
		level = slog.LevelDebug
	}

	writer := os.Stderr
	handler := &LogHandler{
		writer: writer,
		original: slog.NewTextHandler(
			writer,
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
	}

	slog.SetDefault(slog.New(handler))
	slog.SetLogLoggerLevel(slog.LevelInfo) // Set default as info (mostly when using log.Println, etc.)
}

func (handler *LogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return handler.original.Enabled(ctx, level)
}

func (handler *LogHandler) Handle(ctx context.Context, record slog.Record) error {
	var err error
	if record.Level != slog.LevelDebug && record.Level != slog.LevelWarn && record.Level != slog.LevelError {
		_, err = fmt.Fprint(handler.writer, record.Message)
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

		_, err = fmt.Fprintf(handler.writer, "::%s::%s", level, record.Message)
	}

	if err != nil {
		return err //nolint:wrapcheck // Current logger is just a proxy
	}

	// Delegate potential attributes and end of line management to the original handler
	return handler.original.Handle(ctx, record) //nolint:wrapcheck // Current logger is just a proxy
}

func (handler *LogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return handler.original.WithAttrs(attrs)
}

func (handler *LogHandler) WithGroup(name string) slog.Handler {
	return handler.original.WithGroup(name)
}
