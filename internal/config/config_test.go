package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	f := filepath.Join(t.TempDir(), "nonexistent.ini")
	cfg := Load(f)

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 10125, cfg.Port)
	assert.Equal(t, "changeme", cfg.Password)
	assert.FileExists(t, f)
}

func TestLoad_Full(t *testing.T) {
	f := writeTemp(t, `[server]
host = 127.0.0.1
port = 8080

[auth]
password = secret123
`)

	cfg := Load(f)

	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "secret123", cfg.Password)
}

func TestLoad_Partial(t *testing.T) {
	f := writeTemp(t, `[server]
port = 9090
`)

	cfg := Load(f)

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "changeme", cfg.Password)
}

func TestLoad_Comments(t *testing.T) {
	f := writeTemp(t, `# comment
; another comment
[server]
host = 10.0.0.1
`)

	cfg := Load(f)

	assert.Equal(t, "10.0.0.1", cfg.Host)
}

func TestLoad_EmptyFile(t *testing.T) {
	f := writeTemp(t, "")
	cfg := Load(f)

	assert.Equal(t, "0.0.0.0", cfg.Host)
}

func TestLoad_InvalidPort(t *testing.T) {
	f := writeTemp(t, `[server]
port = notanumber
`)

	cfg := Load(f)

	assert.Equal(t, 10125, cfg.Port)
}

func TestLoad_MalformedLines(t *testing.T) {
	f := writeTemp(t, `[server]
host = 127.0.0.1
invalidline
port = 8080
= novalue
`)

	cfg := Load(f)

	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "test.ini")
	os.WriteFile(f, []byte(content), 0644)
	return f
}
