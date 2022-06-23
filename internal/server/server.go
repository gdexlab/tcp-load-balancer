package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"tcp-load-balancer/internal/upstream"
)

// LoadBalancer is a TCP load balancer with methods for handling connections from clients to hosts.
type LoadBalancer struct {
	// listener is the TCP listener for this load balancer.
	listener *net.TCPListener

	// hosts is the list of upstream hosts
	hosts []*upstream.TcpHost

	// hostMu protects the hosts list from concurrent access.
	hostMu sync.RWMutex

	// hostTimeout controls how long the LB will wait for a response from the host prior to timing out.
	hostTimeout time.Duration
}

// Hosts returns the list of hosts that are being load balanced.
func (l *LoadBalancer) Hosts() []*upstream.TcpHost {
	l.hostMu.RLock()
	defer l.hostMu.RUnlock()
	return l.hosts
}

// Address returns the address of the load balancer.
// If the listener is nil, a blank address is returned.
func (l *LoadBalancer) Address() net.Addr {
	if l.listener == nil {
		return &net.TCPAddr{}
	}
	return l.listener.Addr()
}

// AddUpstream adds a new upstream host to the load balancer.
func (l *LoadBalancer) AddUpstream(host *upstream.TcpHost) {
	if host != nil {
		l.hostMu.Lock()
		l.hosts = append(l.hosts, host)
		l.hostMu.Unlock()
	}
}

// New initializes a new LoadBalancer and begins listening for connections.
// Pass :0" as the address to have the load balancer listen on a random port.
func New(tcpNetwork, address string, hostTimeout time.Duration) (*LoadBalancer, error) {
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
		listener:    ln,
		hostTimeout: hostTimeout,
	}, nil
}
