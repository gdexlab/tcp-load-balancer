package server

import (
	"testing"

	"tcp-load-balancer/internal/upstream"
)

func TestLoadBalancer_LeastConnections(t *testing.T) {
	tests := []struct {
		name                    string
		hosts                   []*upstream.TcpHost
		expectedConnectionCount int
		wantErr                 bool
	}{
		{
			name: "host with fewest connections is selected",
			hosts: []*upstream.TcpHost{
				createHostWithNConnections(t, 2),
				// The host with 1 connection should be selected.
				createHostWithNConnections(t, 1),
				createHostWithNConnections(t, 99),
			},
			wantErr:                 false,
			expectedConnectionCount: 1,
		},
		{
			name:    "no hosts returns an error (and does not panic)",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LoadBalancer{
				hosts: tt.hosts,
			}
			got, err := l.LeastConnections()

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadBalancer.LeastConnections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.ConnectionCount() != tt.expectedConnectionCount {
				t.Errorf("LoadBalancer.LeastConnections() expected host with %d connections, got host with %d connections.", tt.expectedConnectionCount, got.ConnectionCount())
			}
		})
	}
}

// createHostWithNConnections creates a host with n active connections.
func createHostWithNConnections(t *testing.T, n int) *upstream.TcpHost {
	h, err := upstream.New(":0", "tcp", n)
	if err != nil {
		t.Error(err)
	}
	for i := 0; i < n; i++ {
		h.IncrementActiveConnections()
	}
	return h
}
