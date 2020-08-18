package log

import (
	"github.com/rs/zerolog"
	"os"
	"time"
)

var Logger zerolog.Logger

func init() {
	Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Caller().Timestamp().Logger()
}
