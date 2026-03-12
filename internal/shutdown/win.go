//go:build windows

package shutdown

import (
	"syscall"
	"unsafe"
)

var (
	user32Dll   = syscall.NewLazyDLL("user32.dll")
	procExitWin = user32Dll.NewProc("ExitWindowsEx")

	advapi32Dll               = syscall.NewLazyDLL("advapi32.dll")
	procOpenProcessToken      = advapi32Dll.NewProc("OpenProcessToken")
	procLookupPrivilegeValue  = advapi32Dll.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges = advapi32Dll.NewProc("AdjustTokenPrivileges")
)

const (
	SePrivilegeEnabled    = 0x00000002
	TokenAdjustPrivileges = 0x0020
	TokenQuery            = 0x0008
	SeShutdownName        = "SeShutdownPrivilege"

	EwxShutdown      = 0x00000001
	EwxForceIfHung   = 0x00000010
	ShtdnReasonOther = 0x00000000
)

type LUID struct {
	LowPart  uint32
	HighPart int32
}

type LuidAndAttributes struct {
	Luid       LUID
	Attributes uint32
}

type TokenPrivileges struct {
	PrivilegeCount uint32
	Privileges     [1]LuidAndAttributes
}

func enablePrivilege() error {
	var token syscall.Token

	pseudoHandle, err := syscall.GetCurrentProcess()
	if err != nil {
		return err
	}

	r1, _, err := procOpenProcessToken.Call(
		uintptr(pseudoHandle),
		uintptr(TokenAdjustPrivileges|TokenQuery),
		uintptr(unsafe.Pointer(&token)),
	)
	if r1 == 0 {
		return err
	}
	defer syscall.CloseHandle(syscall.Handle(token))

	var luid LUID
	namePtr, err := syscall.UTF16PtrFromString(SeShutdownName)
	if err != nil {
		return err
	}

	r1, _, err = procLookupPrivilegeValue.Call(
		0,
		uintptr(unsafe.Pointer(namePtr)),
		uintptr(unsafe.Pointer(&luid)),
	)
	if r1 == 0 {
		return err
	}

	tp := TokenPrivileges{
		PrivilegeCount: 1,
		Privileges: [1]LuidAndAttributes{{
			Luid:       luid,
			Attributes: SePrivilegeEnabled,
		}},
	}

	r1, _, err = procAdjustTokenPrivileges.Call(
		uintptr(token),
		0,
		uintptr(unsafe.Pointer(&tp)),
		0,
		0,
		0,
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func shutdownWithFlags(flags uint32) error {
	if err := enablePrivilege(); err != nil {
		return err
	}
	r1, _, err := procExitWin.Call(
		uintptr(flags),
		uintptr(ShtdnReasonOther),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

func Graceful() error {
	return shutdownWithFlags(EwxShutdown)
}

func ForceIfHung() error {
	return shutdownWithFlags(EwxShutdown | EwxForceIfHung)
}
