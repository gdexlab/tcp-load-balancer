package server

import (
	"errors"

	"tcp-load-balancer/internal/upstream"

	"github.com/google/uuid"
)

var ErrNoHosts = errors.New("no healthy upstream hosts available")

// LeastConnections returns an authorized host with the fewest open connections.
func (l *LoadBalancer) LeastConnections() (*upstream.TcpHost, error) {

	var host *upstream.TcpHost
	for _, h := range l.hosts {
		// TODO: handle authorization scheme here in next PR. For now, we'll just assume the client can access all hosts.

		// If the host's health status is untraceable, or it is unhealthy, skip it.
		if h.ID() == uuid.Nil || l.unhealthyHosts.IsUnhealthy(h.ID()) {
			continue
		}

		// Pick the host with the fewest connections.
		// The first time through this loop, host.ConnectionCount will be 0 because host is nil.
		if host == nil || h.ConnectionCount() < host.ConnectionCount() {
			host = h
		}

	}

	// If the host doesn't have an id, it's health cannot tracked and therefore it cannot be used.
	if host == nil {
		return nil, ErrNoHosts
	}

	return host, nil
}
