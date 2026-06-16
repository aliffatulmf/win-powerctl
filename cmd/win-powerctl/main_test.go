//go:build windows

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRootCmd(t *testing.T) {
	assert.NotNil(t, rootCmd)
	assert.Equal(t, "win-powerctl", rootCmd.Use)
}

func TestVersionCmd(t *testing.T) {
	assert.NotNil(t, versionCmd)
	assert.Equal(t, "version", versionCmd.Use)
}

func TestSubCommands(t *testing.T) {
	commands := rootCmd.Commands()
	names := make([]string, len(commands))
	for i, cmd := range commands {
		names[i] = cmd.Name()
	}

	assert.Contains(t, names, "version")
	assert.Contains(t, names, "install")
	assert.Contains(t, names, "uninstall")
	assert.Contains(t, names, "start")
	assert.Contains(t, names, "stop")
	assert.Contains(t, names, "restart")
	assert.Contains(t, names, "service")
}
