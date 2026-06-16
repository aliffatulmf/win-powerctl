//go:build windows

package poweroff

import (
	"fmt"
	"syscall"
)

var (
	dll          *syscall.DLL
	procGraceful *syscall.Proc
	procForce    *syscall.Proc
	procReboot   *syscall.Proc
	procPowerOff *syscall.Proc
	dllErr       error
)

func init() {
	dll, dllErr = syscall.LoadDLL("poweroff.dll")
	if dllErr != nil {
		return
	}
	procGraceful, _ = dll.FindProc("ShutdownGraceful")
	procForce, _ = dll.FindProc("ShutdownForce")
	procReboot, _ = dll.FindProc("ShutdownReboot")
	procPowerOff, _ = dll.FindProc("ShutdownPowerOff")
}

func callProc(proc *syscall.Proc, dryRun bool) error {
	if proc == nil {
		return fmt.Errorf("poweroff.dll not loaded")
	}
	dr := uintptr(0)
	if dryRun {
		dr = 1
	}
	r1, _, e := proc.Call(dr)
	if r1 == 0 {
		return fmt.Errorf("%s: %w", proc.Name, e)
	}
	return nil
}

func Graceful() error { return callProc(procGraceful, false) }
func Force() error    { return callProc(procForce, false) }
func Reboot() error   { return callProc(procReboot, false) }
func PowerOff() error { return callProc(procPowerOff, false) }

func CheckDLL() error {
	if dllErr != nil {
		return fmt.Errorf("poweroff.dll: %w", dllErr)
	}
	if procGraceful == nil {
		return fmt.Errorf("poweroff.dll: ShutdownGraceful not found")
	}
	return callProc(procGraceful, true)
}
