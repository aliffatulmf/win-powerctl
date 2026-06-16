//go:build windows

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"win-powerctl/internal/config"
)

func newTestServer() *Server {
	cfg := &config.Config{
		Host:     "127.0.0.1",
		Port:     0,
		Password: "testpass",
	}
	s := New(cfg)
	s.SetShutdown(func() error { return nil })
	s.SetCheckDLL(func() error { return nil })
	return s
}

func TestHandleHealth_OK(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	var resp healthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "ok", resp.Status)
	assert.Equal(t, "127.0.0.1", resp.Server.Host)
	assert.Equal(t, 0, resp.Server.Port)
	assert.True(t, resp.Poweroff.Loaded)
}

func TestHandleHealth_DLLFail(t *testing.T) {
	s := newTestServer()
	s.SetCheckDLL(func() error { return assert.AnError })

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp healthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "degraded", resp.Status)
	assert.False(t, resp.Poweroff.Loaded)
	assert.NotEmpty(t, resp.Poweroff.Error)
}

func TestHandleHealth_NoCheckFunc(t *testing.T) {
	cfg := &config.Config{Host: "127.0.0.1", Port: 0, Password: "test"}
	s := New(cfg)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.handleHealth(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp healthResponse
	json.NewDecoder(w.Body).Decode(&resp)
	assert.Equal(t, "degraded", resp.Status)
}

func TestHandleShutdown(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		auth     string
		password string
		wantCode int
		wantBody string
	}{
		{
			name:     "valid password",
			method:   "GET",
			auth:     "testpass",
			password: "testpass",
			wantCode: http.StatusOK,
			wantBody: "Graceful shutdown initiated",
		},
		{
			name:     "wrong password",
			method:   "GET",
			auth:     "wrong",
			password: "testpass",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "no password",
			method:   "GET",
			auth:     "",
			password: "testpass",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "empty password config",
			method:   "GET",
			auth:     "",
			password: "",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:     "method not allowed",
			method:   "POST",
			auth:     "testpass",
			password: "testpass",
			wantCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Host:     "127.0.0.1",
				Port:     0,
				Password: tt.password,
			}
			s := New(cfg)
			s.SetShutdown(func() error { return nil })
			s.SetCheckDLL(func() error { return nil })

			url := "/shutdown"
			if tt.auth != "" {
				url += "?auth=" + tt.auth
			}
			req := httptest.NewRequest(tt.method, url, nil)
			w := httptest.NewRecorder()

			s.handleShutdown(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantBody != "" {
				assert.Contains(t, w.Body.String(), tt.wantBody)
			}

			time.Sleep(10 * time.Millisecond)
		})
	}
}
