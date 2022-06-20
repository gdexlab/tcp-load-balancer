package upstream

import (
	"errors"
	"fmt"
	"log"
	"net"

	"tcp-load-balancer/internal/upstream/connections"
	"tcp-load-balancer/internal/upstream/health"

	"github.com/google/uuid"
)

var (
	ErrUninitialized = errors.New("host uninitialized")
	ErrNoAddress     = errors.New("no upstream host address available")
	ErrUnhealthy     = errors.New("host is unhealthy")
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
	activeConnections connections.Counter

	// healthStatus tracks the number of consecutive failed connection attempts for this host, and returns whether or not it is healthy.
	healthStatus *health.Tracker
}

// IncrementActiveConnections increments the active connection count for this host.
func (h *TcpHost) IncrementActiveConnections() error {
	if h == nil {
		return ErrUninitialized
	}

	h.activeConnections.Increment()
	return nil
}

// DecrementActiveConnections decrements the active connection count for this host.
func (h *TcpHost) DecrementActiveConnections() error {
	if h == nil {
		return ErrUninitialized
	}

	h.activeConnections.Increment()
	return nil
}

// ID returns the id of the host.
func (h *TcpHost) ID() uuid.UUID {
	if h == nil {
		return uuid.Nil
	}
	return h.id
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

	return h.activeConnections.Count()
}

// ShowsHealthy returns true if the consecutive failed connection count is below the failuresThreshold.
// It does not attempt a new connection. If an updated health status needs to be checked, h.Dial can be used.
func (h *TcpHost) ShowsHealthy() bool {
	if h == nil || h.healthStatus == nil {
		return false
	}

	return h.healthStatus.ShowsHealthy()
}

// Dial returns a net connection to the tcp host.
func (h *TcpHost) Dial() (net.Conn, error) {
	if h == nil || h.Address() == nil {
		log.Print("Host or address was nil when dialed; this likely means you need to properly initialize the host.")
		// If the host is nil or has no address, future calls will also fail, so we should return the unhealthy error.
		return nil, ErrUnhealthy
	}

	conn, err := net.Dial(h.network, h.Address().String())
	if err != nil {
		log.Printf("error dialing host: %s", err)

		h.healthStatus.TrackFailure()
		if !h.healthStatus.ShowsHealthy() {
			return nil, ErrUnhealthy
		}
		return nil, err
	}

	h.healthStatus.TrackSuccess()
	return conn, nil
}

// New initializes a new TcpUpstreamHost.
func New(address string, network string, failuresThreshold int) (*TcpHost, error) {
	a, err := net.ResolveTCPAddr(network, address)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve TCP address: %s", err)
	}

	h := health.New(failuresThreshold)

	return &TcpHost{
		address:      a,
		network:      network,
		healthStatus: h,
		id:           uuid.New(),
	}, nil
}
