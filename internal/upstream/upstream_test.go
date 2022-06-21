package upstream

import (
	"testing"
)

func TestTcpHost_DecrementActiveConnections(t *testing.T) {
	tests := []struct {
		name              string
		activeConnections uint64
	}{
		{
			name:              "decrement active connections from 10 to 9",
			activeConnections: 10,
		},
		{
			name:              "active connections never goes below 0",
			activeConnections: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &TcpHost{
				activeConnections: tt.activeConnections,
			}
			h.DecrementActiveConnections()
			if h.activeConnections != tt.activeConnections-1 {
				t.Errorf("TcpHost.DecrementActiveConnections() = %v, want %v", h.activeConnections, tt.activeConnections-1)
			}
		})
	}
}
