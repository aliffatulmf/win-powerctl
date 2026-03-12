//go:build windows

package admin

import (
	"golang.org/x/sys/windows"

	"win-powerctl/internal/logger"
)

func IsElevated() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		logger.Warn("admin", "failed to open process token", "error", err)
		return false
	}
	defer token.Close()

	adminSID, err := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	if err != nil {
		logger.Warn("admin", "failed to create admin SID", "error", err)
		return false
	}

	isMember, err := token.IsMember(adminSID)
	if err != nil {
		logger.Warn("admin", "failed to check token membership", "error", err)
		return false
	}
	return isMember
}
