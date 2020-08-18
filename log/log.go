package log

import (
	"github.com/rs/zerolog"
	"os"
)

var Logger zerolog.Logger

func init() {
	Logger = zerolog.New(os.Stdout).With().Caller().Timestamp().Logger()
}

func Level(level int8) {
	Logger.Level(zerolog.Level(level))
}