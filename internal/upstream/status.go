package upstream

import "sync"

// The load balancer removes unhealthy upstreams if it is unable to connect to the host while handling a client connection.
// The load balancer will store a registry of all hosts and with health status for each host.
// The health checks will allow N consecutive failed attempts before marking unhealthy.
// When an attempt fails to connect, the host will increase `consecutiveFailuresCount`, and the LB will select a different host for the connection.
// If a connection succeeds, the host's `consecutiveFailuresCount` will reset to 0. If N consecutive failed attempts are reached, the host will be marked unhealthy.

// A separate go routing will periodically recheck unhealthy hosts so that statuses can be updated when health is restored.
// A host which allows successful connection will be considered healthy.

// Status represents the status of a host.
type Status int

const (
	StatusUnknown = Status(iota)
	StatusHealthy
	StatusUnhealthy
)

type Health struct {
	// health is the current status of the host.
	status Status

	// statusLock is the current status of the host.
	statusLock sync.Mutex

	// consecutiveFailures is the number of times the host has failed without any successful connection.
	// It will be reset to 0 when a successful connection is made.
	consecutiveFailures int

	// consecutiveFailuresLock ensures that the count of consecutive failures is safe for concurrency.
	consecutiveFailuresLock sync.Mutex
}
