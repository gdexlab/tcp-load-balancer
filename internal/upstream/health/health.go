package health

import "sync"

// TODO: A separate go routing will periodically recheck unhealthy hosts so that statuses can be updated when health is restored.

// Tracker tracks and makes available the health status of a host.
type Tracker struct {
	sync.Mutex

	// consecutiveFailures is the number of times the host has failed without any successful connection.
	// It will be reset to 0 when a successful connection is made.
	consecutiveFailures int

	// failuresThreshold is the number of consecutive failures before the host is considered unhealthy.
	// If the amount of consecutive failures is less than or equal to this threshold, the host will considered healthy.
	failuresThreshold int
}

// TrackFailure increments the consecutive failure count for this host.
func (f *Tracker) TrackFailure() {
	f.Lock()
	f.consecutiveFailures++
	f.Unlock()
}

// TrackSuccess resets the consecutive failure count for this host to 0.
func (f *Tracker) TrackSuccess() {
	f.Lock()
	f.consecutiveFailures = 0
	f.Unlock()
}

// ShowsHealthy returns whether the host shows healthy or not, based on the most recent failed connection counts and the failuresThreshold.
func (f *Tracker) ShowsHealthy() bool {
	f.Lock()
	defer f.Unlock()
	return f.consecutiveFailures <= f.failuresThreshold
}

func New(failuresThreshold int) *Tracker {
	return &Tracker{
		failuresThreshold: failuresThreshold,
	}
}
