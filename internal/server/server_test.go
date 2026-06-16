//go:build windows

package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

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

func TestHandleShutdown_MethodNotAllowed(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("POST", "/shutdown", nil)
	w := httptest.NewRecorder()

	s.handleShutdown(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleShutdown_InvalidPassword(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/shutdown?auth=wrong", nil)
	w := httptest.NewRecorder()

	s.handleShutdown(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleShutdown_ValidPassword(t *testing.T) {
	var called atomic.Bool
	s := newTestServer()
	s.SetShutdown(func() error {
		called.Store(true)
		return nil
	})

	req := httptest.NewRequest("GET", "/shutdown?auth=testpass", nil)
	w := httptest.NewRecorder()

	s.handleShutdown(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Graceful shutdown initiated")
}

func TestHandleShutdown_NoPassword(t *testing.T) {
	s := newTestServer()
	req := httptest.NewRequest("GET", "/shutdown", nil)
	w := httptest.NewRecorder()

	s.handleShutdown(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
