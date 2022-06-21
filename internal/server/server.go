package server

import (
	"errors"
	"fmt"
	"net"

	"tcp-load-balancer/internal/server/unhealthy"
	"tcp-load-balancer/internal/upstream"

	"github.com/google/uuid"
)

// LoadBalancer is a TCP load balancer with methods for handling connections from clients to hosts.
type LoadBalancer struct {
	// listener is the TCP listener for this load balancer.
	listener *net.TCPListener

	// hosts is the list of upstream hosts that are ready for connection.
	// TODO: persist registered hosts in a more permanent data store (outside scope of this project).
	hosts []*upstream.TcpHost

	// unhealthyHosts tracks the upstream hosts that are currently unhealthy.
	unhealthyHosts unhealthy.Hosts
}

// Address returns the address of the load balancer.
func (l *LoadBalancer) Address() net.Addr {
	return l.listener.Addr()
}

// AddUpstream adds a new upstream host to the load balancer, based on the host's ID.
// This method is not safe for concurrent use.
func (l *LoadBalancer) AddUpstream(host *upstream.TcpHost) error {
	if l == nil || host == nil || host.ID() == uuid.Nil {
		return errors.New("unable to add upstream host: host is nil or has no ID")
	}

	l.hosts = append(l.hosts, host)

	return nil
}

// TrackUnhealthyHost adds the host to the unhealthy hosts map.
func (l *LoadBalancer) TrackUnhealthyHost(hostID uuid.UUID) {
	l.unhealthyHosts.Add(hostID)
}

// TrackHealthyHost removes the host from the unhealthy hosts map.
func (l *LoadBalancer) TrackHealthyHost(hostID uuid.UUID) {
	l.unhealthyHosts.Remove(hostID)
}

// New initializes a new LoadBalancer and begins listening for connections.
// Pass :0" as the address to have the load balancer listen on a random port.
func New(tcpNetwork, address string) (*LoadBalancer, error) {
	a, err := net.ResolveTCPAddr(tcpNetwork, address)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve TCP address: %s", err)
	}

	ln, err := net.ListenTCP(tcpNetwork, a)
	if err != nil {
		return nil, fmt.Errorf("unable to listen on %s: %s", a.String(), err)
	}

	// TODO: Load TLS config as part of New LB setup in the PR that handles mTLS requirement.

	return &LoadBalancer{
		listener: ln,
		hosts:    []*upstream.TcpHost{},
	}, nil
}
