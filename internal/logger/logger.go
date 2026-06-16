package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func Init() {
	os.MkdirAll("logs", 0755)

	filename := fmt.Sprintf("logs/%s.log", time.Now().Format("2006-01-02"))
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open log file: %v\n", err)
		os.Exit(1)
	}

	console := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	file := zerolog.ConsoleWriter{Out: f, TimeFormat: "15:04:05", NoColor: true}

	multi := zerolog.MultiLevelWriter(console, file)
	log = zerolog.New(multi).With().Timestamp().Logger()
}

func Info() *zerolog.Event  { return log.Info() }
func Warn() *zerolog.Event  { return log.Warn() }
func Error() *zerolog.Event { return log.Error() }
func Fatal() *zerolog.Event { return log.Fatal() }
func Debug() *zerolog.Event { return log.Debug() }
