package logger

import (
	"os"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

var once sync.Once

var log zerolog.Logger

func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil {
			logLevel = int(zerolog.InfoLevel) // default to INFO
		}

		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		var goVersion string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			goVersion = buildInfo.GoVersion
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)). //nolint:gosec // disable G115
			With().
			Timestamp().
			Str("go_version", goVersion).
			Logger()
	})

	return log
}
