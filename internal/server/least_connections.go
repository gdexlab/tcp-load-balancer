package server

import (
	"errors"

	"tcp-load-balancer/internal/upstream"
)

// LeastConnections returns an authorized host with the fewest open connections.
func (l *LoadBalancer) LeastConnections() (*upstream.TcpHost, error) {
	if l == nil || len(l.hosts) == 0 {
		return nil, errors.New("no upstream hosts available")
	}

	var host *upstream.TcpHost

	for _, h := range l.hosts {
		// TODO: handle authorization scheme here in next PR. For now, we'll just assume the client can access all hosts.
		if host == nil || h.ConnectionCount() < host.ConnectionCount() {
			host = h
		}
	}

	return host, nil
}
