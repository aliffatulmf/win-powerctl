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
