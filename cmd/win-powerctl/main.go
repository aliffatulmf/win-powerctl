package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/spf13/cobra"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"

	"win-powerctl/internal/shutdown"
	"win-powerctl/internal/watchdog"
)

const (
	serviceName = "winpowerctl"
	host        = "0.0.0.0"
	port        = 10125
)

var rootCmd = &cobra.Command{
	Use:   "win-powerctl",
	Short: "Win Power Control Service",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	isService, err := svc.IsWindowsService()
	if err == nil && isService {
		runService()
		return
	}

	rootCmd.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "Install Windows service",
			Run: func(cmd *cobra.Command, args []string) {
				installService()
			},
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Uninstall Windows service",
			Run: func(cmd *cobra.Command, args []string) {
				uninstallService()
			},
		},
		&cobra.Command{
			Use:   "service",
			Short: "Run as Windows service",
			Run: func(cmd *cobra.Command, args []string) {
				runService()
			},
			Hidden: true,
		},
	)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command execution failed: %v", err)
	}
}

func runHTTP(stop <-chan struct{}, errCh chan<- error) {
	r := chi.NewRouter()
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", host, port),
		Handler: r,
	}

	r.Get("/shutdown", func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write([]byte("Graceful shutdown initiated")); err != nil {
			log.Println("write response error:", err)
			return
		}

		go func() {
			watchdog.Start(60 * time.Second)
			if err := shutdown.Graceful(); err != nil {
				log.Println("graceful shutdown error:", err)
			}
		}()
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Println("write response error:", err)
		}
	})

	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Println("http shutdown error:", err)
		}
		close(errCh)
	}()

	log.Printf("Listening on %s:%d", host, port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		select {
		case errCh <- err:
		default:
		}
	}
}

type winsvc struct{}

func (m *winsvc) Execute(_ []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const accepted = svc.AcceptStop | svc.AcceptShutdown
	s <- svc.Status{State: svc.StartPending}

	stopHTTP := make(chan struct{})
	httpErr := make(chan error, 1)

	go func() {
		runHTTP(stopHTTP, httpErr)
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
			default:
			}
		case err := <-httpErr:
			log.Println("http server error:", err)
			s <- svc.Status{State: svc.StopPending}
			close(stopHTTP)
			return false, 1
		}
	}
}

func runService() {
	if err := svc.Run(serviceName, &winsvc{}); err != nil {
		log.Fatalf("service failed: %v", err)
	}
}

func runServiceSC() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd, err := exec.CommandContext(ctx, "sc", "start", serviceName).CombinedOutput()
	if err != nil {
		return fmt.Errorf("sc start command failed: %w, output: %s", err, cmd)
	}
	return nil
}

func stopService(serviceName string, timeout time.Duration) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("connect to SCM: %w", err)
	}
	defer func() {
		if err := m.Disconnect(); err != nil {
			log.Println("SCM disconnect error:", err)
		}
	}()

	s, err := m.OpenService(serviceName)
	if err != nil {
		return fmt.Errorf("open service: %w", err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Println("service close error:", err)
		}
	}()

	status, err := s.Query()
	if err != nil {
		return fmt.Errorf("query service status: %w", err)
	}

	if status.State == svc.Stopped {
		return nil
	}

	if status.State != svc.Running {
		return fmt.Errorf("service not running (current state: %v)", status.State)
	}

	_, err = s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("send stop control: %w", err)
	}

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("query during stop: %w", err)
		}

		if status.State == svc.Stopped {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for service to stop")
}

const (
	firewallName       = "Win Power Control Service"
	NetFwActionAllow   = 1
	NetFwRuleDirIn     = 1
	NetFwProfileAll    = 0x7fffffff
	NetFwIpProtocolAny = 256
)

func withFwPolicy2(f func(policy *ole.IDispatch, rules *ole.IDispatch) error) error {
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		return fmt.Errorf("initialize COM: %w", err)
	}
	defer ole.CoUninitialize()

	policyObj, err := oleutil.CreateObject("HNetCfg.FwPolicy2")
	if err != nil {
		return fmt.Errorf("create FwPolicy2: %w", err)
	}
	defer policyObj.Release()

	policy, err := policyObj.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("query IDispatch: %w", err)
	}
	defer policy.Release()

	rulesRaw, err := oleutil.GetProperty(policy, "Rules")
	if err != nil {
		return fmt.Errorf("get Rules: %w", err)
	}
	rules := rulesRaw.ToIDispatch()
	defer rules.Release()

	return f(policy, rules)
}

func installService() {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	m, err := mgr.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := m.Disconnect(); err != nil {
			log.Println("SCM disconnect error:", err)
		}
	}()

	s, err := m.OpenService(serviceName)
	if err == nil {
		if err := s.Close(); err != nil {
			log.Println("service close error:", err)
		}
		log.Println("service already installed")
		return
	}

	s, err = m.CreateService(serviceName, exe, mgr.Config{
		DisplayName: "Win Power Control",
		StartType:   mgr.StartAutomatic,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Println("service close error:", err)
		}
	}()

	log.Println("service installed")

	if err := runServiceSC(); err != nil {
		log.Printf("failed to start service: %v", err)
	}

	log.Println("service started")

	if err := excludeFromFirewall(); err != nil {
		log.Printf("failed to exclude from firewall: %v", err)
	}

	log.Println("firewall rule added")
}

func uninstallService() {
	m, err := mgr.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := m.Disconnect(); err != nil {
			log.Println("SCM disconnect error:", err)
		}
	}()

	s, err := m.OpenService(serviceName)
	if err != nil {
		log.Fatal("service not installed")
	}
	defer func() {
		if err := s.Close(); err != nil {
			log.Println("service close error:", err)
		}
	}()

	if err := stopService(serviceName, 10*time.Second); err != nil {
		log.Printf("failed to stop service: %v", err)
	} else {
		log.Println("service stopped")
	}
	if err := s.Delete(); err != nil {
		log.Fatal(err)
	}

	log.Println("service deleted")

	if err := removeFirewallRule(); err != nil {
		log.Println("failed to remove firewall rule")
	} else {
		log.Println("firewall rule removed")
	}
}

func excludeFromFirewall() error {
	appPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("resolve executable path: %w", err)
	}

	return withFwPolicy2(func(policy, rules *ole.IDispatch) error {
		_, _ = oleutil.CallMethod(rules, "Remove", firewallName)

		ruleObj, err := oleutil.CreateObject("HNetCfg.FWRule")
		if err != nil {
			return fmt.Errorf("create FWRule: %w", err)
		}
		defer ruleObj.Release()

		rule, err := ruleObj.QueryInterface(ole.IID_IDispatch)
		if err != nil {
			return fmt.Errorf("query FWRule IDispatch: %w", err)
		}
		defer rule.Release()

		if _, err = oleutil.PutProperty(rule, "Name", firewallName); err != nil {
			return err
		}
		if _, err = oleutil.PutProperty(rule, "ApplicationName", appPath); err != nil {
			return err
		}
		if _, err = oleutil.PutProperty(rule, "Action", NetFwActionAllow); err != nil {
			return err
		}
		if _, err = oleutil.PutProperty(rule, "Direction", NetFwRuleDirIn); err != nil {
			return err
		}
		if _, err = oleutil.PutProperty(rule, "Enabled", true); err != nil {
			return err
		}
		if _, err = oleutil.PutProperty(rule, "Profiles", NetFwProfileAll); err != nil {
			return err
		}
		if _, err = oleutil.PutProperty(rule, "Protocol", NetFwIpProtocolAny); err != nil {
			return err
		}
		if _, err = oleutil.CallMethod(rules, "Add", rule); err != nil {
			return fmt.Errorf("add firewall rule: %w", err)
		}
		return nil
	})
}

func removeFirewallRule() error {
	return withFwPolicy2(func(policy, rules *ole.IDispatch) error {
		_, err := oleutil.CallMethod(rules, "Remove", firewallName)
		if err != nil {
			return fmt.Errorf("firewall rule not found")
		}
		return nil
	})
}
