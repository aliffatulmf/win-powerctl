//go:build windows

package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows"

	"win-powerctl/internal/config"
	"win-powerctl/internal/logger"
	"win-powerctl/internal/poweroff"
	"win-powerctl/internal/server"
	"win-powerctl/internal/service"
)

const (
	serviceName = "winpowerctl"
	version     = "1.0.0"
)

func main() {
	logger.Init()
	cfg := config.Load("config.ini")

	if len(os.Args) < 2 {
		logger.Info().Str("host", cfg.Host).Int("port", cfg.Port).Msg("starting server")
		s := server.New(cfg)
		s.SetShutdown(poweroff.Graceful)
		s.SetCheckDLL(poweroff.CheckDLL)
		service.Run(serviceName, s)
		return
	}

	logger.Info().Str("command", os.Args[1]).Msg("command received")

	switch os.Args[1] {
	case "install":
		requireAdmin()
		service.Install(serviceName)
	case "uninstall":
		requireAdmin()
		service.Uninstall(serviceName)
	case "start":
		requireAdmin()
		service.Start(serviceName)
	case "stop":
		requireAdmin()
		service.Stop(serviceName)
	case "restart":
		requireAdmin()
		service.Restart(serviceName)
	case "service":
		s := server.New(cfg)
		s.SetShutdown(poweroff.Graceful)
		s.SetCheckDLL(poweroff.CheckDLL)
		service.Run(serviceName, s)
	case "version":
		fmt.Printf("win-powerctl %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "usage: win-powerctl [install|uninstall|start|stop|restart|service|version]")
		os.Exit(1)
	}
}

func requireAdmin() {
	if windows.GetCurrentProcessToken().IsElevated() {
		return
	}
	fmt.Fprintln(os.Stderr, "Administrator privileges required")
	os.Exit(1)
}
