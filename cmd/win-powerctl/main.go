//go:build windows

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

var cfg *config.Config

var rootCmd = &cobra.Command{
	Use:   "win-powerctl",
	Short: "HTTP service for remote system power control",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger.Info().Str("host", cfg.Host).Int("port", cfg.Port).Msg("starting server")
		s := server.New(cfg)
		s.SetShutdown(poweroff.Graceful)
		s.SetCheckDLL(poweroff.CheckDLL)
		service.Run(serviceName, s)
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("win-powerctl %s\n", version)
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install as Windows service",
	Run: func(cmd *cobra.Command, args []string) {
		requireAdmin()
		service.Install(serviceName)
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove from Windows services",
	Run: func(cmd *cobra.Command, args []string) {
		requireAdmin()
		service.Uninstall(serviceName)
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the service",
	Run: func(cmd *cobra.Command, args []string) {
		requireAdmin()
		service.Start(serviceName)
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the service",
	Run: func(cmd *cobra.Command, args []string) {
		requireAdmin()
		service.Stop(serviceName)
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the service",
	Run: func(cmd *cobra.Command, args []string) {
		requireAdmin()
		service.Restart(serviceName)
	},
}

var serviceCmd = &cobra.Command{
	Use:    "service",
	Short:  "Run as Windows service",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		s := server.New(cfg)
		s.SetShutdown(poweroff.Graceful)
		s.SetCheckDLL(poweroff.CheckDLL)
		service.Run(serviceName, s)
		return nil
	},
}

func init() {
	cfg = config.Load("config.ini")

	rootCmd.AddCommand(
		versionCmd,
		installCmd,
		uninstallCmd,
		startCmd,
		stopCmd,
		restartCmd,
		serviceCmd,
	)
}

func main() {
	logger.Init()

	if err := rootCmd.Execute(); err != nil {
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
