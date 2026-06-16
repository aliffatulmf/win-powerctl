package logger

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	Init()
	assert.DirExists(t, "logs")

	files, err := os.ReadDir("logs")
	assert.NoError(t, err)
	assert.Greater(t, len(files), 0)
}

func TestLogOutput(t *testing.T) {
	Init()
	Info().Str("key", "value").Msg("test message")
}
