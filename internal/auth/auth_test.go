package auth

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadPassword_Valid(t *testing.T) {
	fname := "test_password.txt"
	pw := "mysecret"
	err := os.WriteFile(fname, []byte(pw), 0600)
	assert.NoError(t, err, "failed to write test file")
	defer os.Remove(fname)

	orig := "password.txt"
	backup := "password_backup.txt"
	if _, err := os.Stat(orig); err == nil {
		os.Rename(orig, backup)
		defer os.Rename(backup, orig)
	}
	os.Rename(fname, orig)
	defer os.Remove(orig)

	got, err := ReadPassword()
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, pw, got, "expected password match")
}

func TestReadPassword_FileMissing(t *testing.T) {
	orig := "password.txt"
	backup := "password_backup.txt"
	if _, err := os.Stat(orig); err == nil {
		os.Rename(orig, backup)
		defer os.Rename(backup, orig)
	}
	os.Remove(orig)

	_, err := ReadPassword()
	assert.ErrorIs(t, err, ErrPasswordFileMissing, "expected ErrPasswordFileMissing")
}

func TestReadPassword_Empty(t *testing.T) {
	orig := "password.txt"
	backup := "password_backup.txt"
	if _, err := os.Stat(orig); err == nil {
		os.Rename(orig, backup)
		defer os.Rename(backup, orig)
	}
	os.WriteFile(orig, []byte(""), 0600)
	defer os.Remove(orig)

	_, err := ReadPassword()
	assert.ErrorIs(t, err, ErrPasswordEmpty, "expected ErrPasswordEmpty")
}
