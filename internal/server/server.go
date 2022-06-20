package server

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"tcp-load-balancer/internal/upstream"

	"github.com/google/uuid"
)

// LoadBalancer is a TCP load balancer with methods for handling connections from clients to hosts.
type LoadBalancer struct {
	// listener is the TCP listener for this load balancer.
	listener *net.TCPListener

	// healthyHosts is the list of upstream hosts that are ready for connection.
	// TODO: persist registered hosts in a more permanent data store (outside scope of this project).
	healthyHosts     map[uuid.UUID]*upstream.TcpHost
	healthyHostsLock sync.Mutex

	// unhealthyHosts is the map of upstream hosts that are currently unhealthy.
	// TODO: persist registered hosts in a more permanent data store (outside scope of this project).
	unhealthyHosts     map[uuid.UUID]*upstream.TcpHost
	unhealthyHostsLock sync.Mutex
}

// Address returns the address of the load balancer.
func (l *LoadBalancer) Address() net.Addr {
	return l.listener.Addr()
}

// AddUpstream adds a new upstream host to the load balancer, based on the host's ID.
func (l *LoadBalancer) AddUpstream(host *upstream.TcpHost) error {
	if l == nil || host == nil || host.ID() == uuid.Nil {
		return errors.New("unable to add upstream host: host is nil or has no ID")
	}

	l.healthyHostsLock.Lock()
	defer l.healthyHostsLock.Unlock()

	// Create the healthy hosts map if it doesn't exist.
	if l.healthyHosts == nil {
		l.healthyHosts = map[uuid.UUID]*upstream.TcpHost{}
	}
	l.healthyHosts[host.ID()] = host

	return nil
}

// MarkHostUnhealthy removes the host from the healthy hosts slice, and adds it into the unhealthy hosts map.
func (l *LoadBalancer) MarkHostUnhealthy(hostID uuid.UUID) {
	l.unhealthyHostsLock.Lock()
	l.healthyHostsLock.Lock()

	l.validationInitialization()
	fmt.Println("host was removed from map!!!!")

	h := l.healthyHosts[hostID]
	l.unhealthyHosts[hostID] = h
	delete(l.healthyHosts, h.ID())

	l.healthyHostsLock.Unlock()
	l.unhealthyHostsLock.Unlock()
}

// MarkHostHealthy removes the host from the healthy hosts slice, and adds it into the unhealthy hosts map.
func (l *LoadBalancer) MarkHostHealthy(hostID uuid.UUID) {
	l.unhealthyHostsLock.Lock()
	l.healthyHostsLock.Lock()

	l.validationInitialization()

	h := l.unhealthyHosts[hostID]
	l.healthyHosts[hostID] = h
	delete(l.unhealthyHosts, h.ID())

	l.unhealthyHostsLock.Unlock()
	l.healthyHostsLock.Unlock()
}

// validationInitialization ensures that maps have been properly initialized. It is not concurrency safe if called by itself. Call with locks held.
func (l *LoadBalancer) validationInitialization() {
	// Create the healthy hosts map if it doesn't exist.
	if l.healthyHosts == nil {
		l.healthyHosts = make(map[uuid.UUID]*upstream.TcpHost)
	}

	// Create the unhealthy hosts map if it doesn't exist.
	if l.unhealthyHosts == nil {
		l.unhealthyHosts = make(map[uuid.UUID]*upstream.TcpHost)
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
		listener:       ln,
		healthyHosts:   map[uuid.UUID]*upstream.TcpHost{},
		unhealthyHosts: map[uuid.UUID]*upstream.TcpHost{},
	}, nil
}
