package upstream

import (
	"errors"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/google/uuid"
)

var (
	ErrNoAddress = errors.New("no upstream host address available")
)

// TcpHost represents the upstream hosts to which the LB connects and forwards data.
type TcpHost struct {
	// id is the unique identifier of this host.
	id uuid.UUID

	// address is the remote address of this upstream host.
	address *net.TCPAddr

	// network is the network type of the TcpHost. One of "tcp", "tcp4", "tcp6"
	network string

	// activeConnections tracks the number of open connections to the host.
	activeConnections uint64
}

// IncrementActiveConnections increments the active connection count for this host.
func (h *TcpHost) IncrementActiveConnections() {
	atomic.AddUint64(&h.activeConnections, uint64(1))
}

// DecrementActiveConnections decrements the active connection count for this host.
func (h *TcpHost) DecrementActiveConnections() {
	atomic.AddUint64(&h.activeConnections, ^uint64(0))
}

// Address returns the address of the host.
func (h *TcpHost) Address() *net.TCPAddr {
	return h.address
}

// ConnectionCount returns the number of active connections to this host.
func (h *TcpHost) ConnectionCount() uint64 {
	return h.activeConnections
}

// Dial returns a net connection to the tcp host.
func (h *TcpHost) Dial() (net.Conn, error) {
	if h.Address() == nil {
		return nil, ErrNoAddress
	}

	return net.Dial(h.network, h.Address().String())
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
