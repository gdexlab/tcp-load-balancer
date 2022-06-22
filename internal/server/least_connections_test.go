package server

import (
	"reflect"
	"testing"

	"tcp-load-balancer/internal/upstream"
)

func TestLoadBalancer_LeastConnections(t *testing.T) {
	tests := []struct {
		name            string
		l               *LoadBalancer
		wantErr         bool
		wantHostAtIndex int
	}{
		{
			name: "host with fewest connections is selected",
			l: func() *LoadBalancer {
				h0 := &upstream.TcpHost{}
				h0.IncrementActiveConnections()
				h0.IncrementActiveConnections()

				// h1 has the fewest connections and should be selected.
				h1 := &upstream.TcpHost{}
				h1.IncrementActiveConnections()

				h2 := &upstream.TcpHost{}
				h2.IncrementActiveConnections()
				h2.IncrementActiveConnections()

				return &LoadBalancer{
					hosts: []*upstream.TcpHost{h0, h1, h2},
				}
			}(),
			wantErr:         false,
			wantHostAtIndex: 1,
		},
		{
			name: "first host in slice is selected when all connection counts are equal",
			l: func() *LoadBalancer {
				h0 := &upstream.TcpHost{}
				h0.IncrementActiveConnections()

				h1 := &upstream.TcpHost{}
				h1.IncrementActiveConnections()

				h2 := &upstream.TcpHost{}
				h2.IncrementActiveConnections()

				return &LoadBalancer{
					hosts: []*upstream.TcpHost{h0, h1, h2},
				}
			}(),
			wantErr:         false,
			wantHostAtIndex: 0,
		},
		{
			name:    "no hosts returns an error (and does not panic)",
			l:       &LoadBalancer{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.LeastConnections()

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadBalancer.LeastConnections() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.l.hosts[tt.wantHostAtIndex]) {
				t.Errorf("LoadBalancer.LeastConnections() = %v, want %v", got, tt.l.hosts[tt.wantHostAtIndex])
			}
		})
	}
}
