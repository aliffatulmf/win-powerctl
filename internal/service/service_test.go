//go:build windows

package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockServer struct {
	called bool
}

func (m *mockServer) Run(stop <-chan struct{}, errCh chan<- error) {
	m.called = true
	<-stop
}

func TestServerInterface(t *testing.T) {
	var s Server = &mockServer{}
	assert.NotNil(t, s)
}

func TestMockServer_Called(t *testing.T) {
	m := &mockServer{}
	stop := make(chan struct{})
	errCh := make(chan error, 1)

	done := make(chan struct{})
	go func() {
		m.Run(stop, errCh)
		close(done)
	}()

	close(stop)
	<-done

	assert.True(t, m.called)
}
