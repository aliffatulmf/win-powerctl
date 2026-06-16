//go:build windows

package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"win-powerctl/internal/config"
	"win-powerctl/internal/logger"
)

type Server struct {
	cfg      *config.Config
	mux      *http.ServeMux
	srv      *http.Server
	once     sync.Once
	stop     chan struct{}
	shutdown func() error
	checkDLL func() error
}

func New(cfg *config.Config) *Server {
	s := &Server{cfg: cfg, stop: make(chan struct{})}
	s.routes()
	return s
}

func (s *Server) SetShutdown(fn func() error) {
	s.shutdown = fn
}

func (s *Server) SetCheckDLL(fn func() error) {
	s.checkDLL = fn
}

func (s *Server) routes() {
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("/shutdown", s.handleShutdown)
	s.mux.HandleFunc("/health", s.handleHealth)
}

func (s *Server) handleShutdown(w http.ResponseWriter, r *http.Request) {
	logger.Info().Str("method", r.Method).Str("remote", r.RemoteAddr).Msg("shutdown request")

	if r.Method != http.MethodGet {
		logger.Warn().Str("method", r.Method).Msg("method not allowed")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	authParam := r.URL.Query().Get("auth")
	if s.cfg.Password == "" || s.cfg.Password != authParam {
		logger.Warn().Str("remote", r.RemoteAddr).Msg("invalid auth")
		http.Error(w, "Invalid authentication", http.StatusUnauthorized)
		return
	}

	logger.Info().Str("remote", r.RemoteAddr).Msg("shutdown initiated")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Write([]byte("Graceful shutdown initiated"))

	s.once.Do(func() {
		go func() {
			logger.Info().Msg("gracefully stopping HTTP server")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			s.srv.Shutdown(ctx)

			logger.Info().Msg("triggering system poweroff")
			s.shutdown()
		}()
	})
}

type healthResponse struct {
	Status   string       `json:"status"`
	Server   serverStatus `json:"server"`
	Poweroff dllStatus    `json:"poweroff"`
}

type serverStatus struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

type dllStatus struct {
	Loaded bool   `json:"loaded"`
	Error  string `json:"error,omitempty"`
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	resp := healthResponse{
		Status: "ok",
		Server: serverStatus{
			Host: s.cfg.Host,
			Port: s.cfg.Port,
		},
	}

	if s.checkDLL != nil {
		if err := s.checkDLL(); err != nil {
			resp.Status = "degraded"
			resp.Poweroff = dllStatus{Loaded: false, Error: err.Error()}
		} else {
			resp.Poweroff = dllStatus{Loaded: true}
		}
	} else {
		resp.Status = "degraded"
		resp.Poweroff = dllStatus{Loaded: false, Error: "check function not configured"}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	statusCode := http.StatusOK
	if resp.Status == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) Run(stop <-chan struct{}, errCh chan<- error) {
	s.srv = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port),
		Handler: s.mux,
	}

	go func() {
		<-stop
		logger.Info().Msg("service stop received, shutting down HTTP server")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.srv.Shutdown(ctx)
		close(errCh)
	}()

	logger.Info().Str("addr", s.srv.Addr).Msg("HTTP server listening")
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error().Err(err).Msg("HTTP server error")
		select {
		case errCh <- err:
		default:
		}
	}
}
