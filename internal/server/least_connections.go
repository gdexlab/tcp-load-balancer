package server

import (
	"errors"

	"tcp-load-balancer/internal/upstream"
)

// LeastConnections returns an authorized host with the fewest open connections.
func (l *LoadBalancer) LeastConnections() (*upstream.TcpHost, error) {
	if len(l.hosts) == 0 {
		return nil, errors.New("no upstream hosts available")
	}

	var selectedHost *upstream.TcpHost

	for _, h := range l.hosts {
		// TODO: handle authorization scheme here in next PR. For now, we'll just assume the client can access all hosts.
		if selectedHost == nil || h.ConnectionCount() < selectedHost.ConnectionCount() {
			selectedHost = h
		}
	}

	return selectedHost, nil
}
