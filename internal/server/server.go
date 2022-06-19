package server

import (
	"fmt"
	"net"

	"tcp-load-balancer/internal/upstream"
)

// LoadBalancer is a TCP load balancer with methods for handling connections from clients to hosts.
type LoadBalancer struct {
	// listener is the TCP listener for this load balancer.
	listener *net.TCPListener

	// hosts is the list of upstream hosts
	hosts []*upstream.TcpHost
}

// Hosts returns the list of hosts that are being load balanced.
func (l *LoadBalancer) Hosts() []*upstream.TcpHost {
	return l.hosts
}

// Address returns the address of the load balancer.
func (l *LoadBalancer) Address() net.Addr {
	return l.listener.Addr()
}

// AddUpstream adds a new upstream host to the load balancer.
func (l *LoadBalancer) AddUpstream(host *upstream.TcpHost) {
	if host != nil {
		l.hosts = append(l.hosts, host)
	}
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
