//go:build windows

package service

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

type Server interface {
	Run(stop <-chan struct{}, errCh chan<- error)
}

type winSvc struct {
	server Server
}

func (m *winSvc) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}

	stopHTTP := make(chan struct{})
	httpErr := make(chan error, 1)

	go func() {
		m.server.Run(stopHTTP, httpErr)
	}()

	s <- svc.Status{State: svc.Running, Accepts: accepted}

	for {
		select {
		case c, ok := <-r:
			if !ok {
				close(stopHTTP)
				return false, 0
			}
			switch c.Cmd {
			case svc.Interrogate:
				s <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s <- svc.Status{State: svc.StopPending}
				close(stopHTTP)
				return false, 0
			}
		case <-httpErr:
			s <- svc.Status{State: svc.StopPending}
			close(stopHTTP)
			return false, 1
		}
	}
}

func Run(name string, server Server) {
	if err := svc.Run(name, &winSvc{server: server}); err != nil {
		fmt.Fprintf(os.Stderr, "service failed: %v\n", err)
		os.Exit(1)
	}
}

func Install(name string) {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "resolve executable: %v\n", err)
		os.Exit(1)
	}

	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to SCM: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return
	}

	s, err = m.CreateService(name, exe, mgr.Config{
		DisplayName:  "Win Power Control",
		Description:  "HTTP service for remote system power control (shutdown, restart, power off)",
		StartType:    mgr.StartAutomatic,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "create service: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()
}

func Uninstall(name string) {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to SCM: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service not installed: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	s.Control(svc.Stop)
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		status, err := s.Query()
		if err == nil && status.State == svc.Stopped {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if err := s.Delete(); err != nil {
		fmt.Fprintf(os.Stderr, "delete service: %v\n", err)
		os.Exit(1)
	}
}

func Start(name string) {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to SCM: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service not installed: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		fmt.Fprintf(os.Stderr, "query service: %v\n", err)
		os.Exit(1)
	}
	if status.State == svc.Running {
		fmt.Println("service already running")
		return
	}

	if err := s.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "start service: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("service started")
}

func Stop(name string) {
	m, err := mgr.Connect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect to SCM: %v\n", err)
		os.Exit(1)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "service not installed: %v\n", err)
		os.Exit(1)
	}
	defer s.Close()

	status, err := s.Query()
	if err != nil {
		fmt.Fprintf(os.Stderr, "query service: %v\n", err)
		os.Exit(1)
	}
	if status.State == svc.Stopped {
		fmt.Println("service already stopped")
		return
	}

	if _, err := s.Control(svc.Stop); err != nil {
		fmt.Fprintf(os.Stderr, "stop service: %v\n", err)
		os.Exit(1)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		status, err := s.Query()
		if err == nil && status.State == svc.Stopped {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("service stopped")
}

func Restart(name string) {
	Stop(name)
	Start(name)
}
