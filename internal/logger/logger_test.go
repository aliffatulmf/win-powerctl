package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	origDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	Init()
	defer Close()

	assert.DirExists(t, filepath.Join(tempDir, "logs"))

	files, err := os.ReadDir(filepath.Join(tempDir, "logs"))
	assert.NoError(t, err)
	assert.Greater(t, len(files), 0)
}

func TestLogOutput(t *testing.T) {
	origDir, _ := os.Getwd()
	tempDir := t.TempDir()
	os.Chdir(tempDir)
	defer os.Chdir(origDir)

	Init()
	defer Close()

	Info().Str("key", "value").Msg("test message")
	Warn().Str("key", "value").Msg("test warning")
	Error().Str("key", "value").Msg("test error")
	Debug().Str("key", "value").Msg("test debug")

	logsDir := filepath.Join(tempDir, "logs")
	files, err := os.ReadDir(logsDir)
	assert.NoError(t, err)
	assert.Greater(t, len(files), 0)
}
