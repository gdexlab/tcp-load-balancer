package server

import (
	"errors"

	"tcp-load-balancer/internal/upstream"
)

// LeastConnections returns an authorized host with the fewest open connections.
func (l *LoadBalancer) LeastConnections() (*upstream.TcpHost, error) {
	if l == nil {
		return nil, errors.New("LB is nil")
	}

	l.healthyHostsLock.Lock()
	defer l.healthyHostsLock.Unlock()

	// TODO: now that we're locking this, it's going to SLOW THINGS DOWN :thinking_face:
	// CONSIDER alternative locking approach, like filtering out hosts that are unhealthy, without locking this thing.
	// Another option is to copy the healthy hosts slice/map, while locked, and then iterate over that copy.

	// The problem is that we update healthy hosts in a routine so we can't expect healthy hosts to be safe.

	if len(l.healthyHosts) == 0 {
		return nil, errors.New("no upstream hosts available")
	}

	var host *upstream.TcpHost

	for _, h := range l.healthyHosts {
		// TODO: handle authorization scheme here in next PR. For now, we'll just assume the client can access all hosts.
		if host == nil || h.ConnectionCount() < host.ConnectionCount() {
			host = h
		}
	}
	return host, nil
}
