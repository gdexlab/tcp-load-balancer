package server

import (
	"reflect"
	"testing"

	"tcp-load-balancer/internal/upstream"

	"github.com/google/uuid"
)

func TestLoadBalancer_LeastConnections(t *testing.T) {
	hostID1 := uuid.New()
	hostID2 := uuid.New()
	hostID3 := uuid.New()

	tests := []struct {
		name         string
		healthyHosts map[uuid.UUID]*upstream.TcpHost
		wantErr      bool
		wantHost     uuid.UUID
	}{
		{
			name: "host with fewest connections is selected",
			healthyHosts: map[uuid.UUID]*upstream.TcpHost{
				hostID1: createHostWithNConnections(t, 2),
				// hostID2 has the fewest connections and should be selected.
				hostID2: createHostWithNConnections(t, 1),
				hostID3: createHostWithNConnections(t, 99),
			},
			wantErr:  false,
			wantHost: hostID2,
		},
		{
			name:    "no hosts returns an error (and does not panic)",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LoadBalancer{
				healthyHosts: tt.healthyHosts,
			}
			got, err := l.LeastConnections()

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadBalancer.LeastConnections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, l.healthyHosts[tt.wantHost]) {
				t.Errorf("LoadBalancer.LeastConnections() = %v, want %v", got, l.healthyHosts[tt.wantHost])
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
