package upstream

import (
	"tcp-load-balancer/internal/upstream/connections"
	"testing"
)

func TestTcpHost_IncrementActiveConnections(t *testing.T) {
	tests := []struct {
		name string
		host *TcpHost
	}{
		{
			name: "Accurately increase connection counter.",
			host: &TcpHost{
				activeConnections: connections.Counter{},
			},
		},
		{
			name: "Increment the active connection count even when the counter is not explicitly initialized.",
			host: &TcpHost{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.host.IncrementActiveConnections()
			if tt.host.ConnectionCount() != 1 {
				t.Errorf("TcpHost.IncrementActiveConnections() did not increment the active connection count.")
			}
		})
	}
}
