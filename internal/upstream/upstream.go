package upstream

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
)

var (
	ErrUninitialized = errors.New("host uninitialized")
	ErrNoAddress     = errors.New("no upstream host address available")
)

// TcpHost represents the upstream hosts to which the LB connects and forwards data.
type TcpHost struct {
	// address is the remote address of this upstream host.
	address *net.TCPAddr

	// network is the network type of the TcpHost. One of "tcp", "tcp4", "tcp6"
	network string

	// activeConnections tracks the number of open connections to the host.
	activeConnections int

	// ID is the unique identifier of this host.
	ID uuid.UUID

	// activeConnectionsLock enables concurrent safety for reading and writing to the activeConnections field.
	activeConnectionsLock sync.Mutex

	// health tra
	health Health
}

// IncrementActiveConnections increments the active connection count for this host.
func (h *TcpHost) IncrementActiveConnections() error {
	if h == nil {
		return ErrUninitialized
	}

	h.connectionLock.Lock()
	h.activeConnections++
	h.connectionLock.Unlock()
	return nil
}

// DecrementActiveConnections decrements the active connection count for this host.
func (h *TcpHost) DecrementActiveConnections() error {
	if h == nil {
		return ErrUninitialized
	}

	h.connectionLock.Lock()
	h.activeConnections--
	h.connectionLock.Unlock()
	return nil
}

// Address returns the address of the host.
func (h *TcpHost) Address() *net.TCPAddr {
	if h == nil {
		return nil
	}
	return h.address
}

// ConnectionCount returns the number of active connections to this host.
func (h *TcpHost) ConnectionCount() int {
	if h == nil {
		return 0
	}

	h.connectionLock.Lock()
	defer h.connectionLock.Unlock()
	return h.activeConnections
}

// Dial returns a net connection to the tcp host.
func (h *TcpHost) Dial() (net.Conn, error) {
	if h == nil {
		return nil, ErrUninitialized
	}

	if h.Address() == nil {
		return nil, ErrNoAddress
	}

	conn, err := net.Dial(h.network, h.Address().String())
	if err != nil {
		h.healthLock.Lock()
		h.health = StatusUnhealthy
		// TODO: increment a counter for the number of times this host has failed
		h.healthLock.Unlock()
		return nil, err
	}

	return conn, nil
}

// Healthy returns true if the health status of the host equals StatusHealthy.
func (h *TcpHost) Healthy() bool {
	if h == nil {
		return false
	}

	h.healthLock.Lock()
	defer h.healthLock.Unlock()
	return h.health == StatusHealthy
}

// New initializes a new TcpUpstreamHost.
func New(address, network string) (*TcpHost, error) {
	a, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve TCP address: %s", err)
	}

	return &TcpHost{
		address: a,
		network: network,
		// TODO: Add hostIDs during PR with authorization scheme
	}, nil
}
