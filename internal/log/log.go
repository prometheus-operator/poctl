package log

import (
	"errors"
	"log/slog"
	"os"

	flag "github.com/spf13/pflag"
)

var (
	logLevel  string
	logFormat string
)

func RegisterFlags(flagSet *flag.FlagSet) {
	flagSet.StringVar(&logLevel, "log-level", slog.LevelDebug.String(), "Log level")
	flagSet.StringVar(&logFormat, "log-format", "text", "Log format")
}

func parseLogLevel() (*slog.Level, error) {
	level := slog.Level(1)
	err := level.UnmarshalText([]byte(logLevel))
	if err != nil {
		return nil, err
	}
	return &level, nil
}

func NewLogger() (*slog.Logger, error) {
	level, err := parseLogLevel()
	if err != nil {
		return nil, err
	}

	var handler slog.Handler
	handlerOptions := &slog.HandlerOptions{
		Level: *level,
	}
	switch {
	case logFormat == "text":
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	case logFormat == "json":
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		return nil, errors.New("unknown log format")
	}
	return slog.New(handler), nil
}
