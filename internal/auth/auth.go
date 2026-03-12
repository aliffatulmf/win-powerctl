package auth

import (
	"errors"
	"os"
	"strings"
)

var (
	ErrPasswordFileMissing = errors.New("password file missing or unreadable")
	ErrPasswordEmpty       = errors.New("password file empty")
)

func ReadPassword() (string, error) {
	data, err := os.ReadFile("password.txt")
	if err != nil {
		return "", ErrPasswordFileMissing
	}
	password := strings.TrimSpace(string(data))
	if password == "" {
		return "", ErrPasswordEmpty
	}
	return password, nil
}
