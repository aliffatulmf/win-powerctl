package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	cfg := Load("nonexistent.ini")
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 10125, cfg.Port)
	assert.Equal(t, "changeme", cfg.Password)
}

func TestLoad_Full(t *testing.T) {
	content := `[server]
host = 127.0.0.1
port = 8080

[auth]
password = secret123
`
	f := "test_config.ini"
	os.WriteFile(f, []byte(content), 0644)
	defer os.Remove(f)

	cfg := Load(f)
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "secret123", cfg.Password)
}

func TestLoad_Partial(t *testing.T) {
	content := `[server]
port = 9090
`
	f := "test_partial.ini"
	os.WriteFile(f, []byte(content), 0644)
	defer os.Remove(f)

	cfg := Load(f)
	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "changeme", cfg.Password)
}

func TestLoad_Comments(t *testing.T) {
	content := `# comment
; another comment
[server]
host = 10.0.0.1
`
	f := "test_comments.ini"
	os.WriteFile(f, []byte(content), 0644)
	defer os.Remove(f)

	cfg := Load(f)
	assert.Equal(t, "10.0.0.1", cfg.Host)
}

func TestLoad_EmptyFile(t *testing.T) {
	f := "test_empty.ini"
	os.WriteFile(f, []byte(""), 0644)
	defer os.Remove(f)

	cfg := Load(f)
	assert.Equal(t, "0.0.0.0", cfg.Host)
}
