package unhealthy

import (
	"sync"

	"github.com/google/uuid"
)

// Hosts is a concurrent-safe map of unhealthy hosts.
type Hosts struct {
	sync.Mutex

	// unhealthyHosts is the set of upstream hosts which are currently unhealthy.
	hostIDs map[uuid.UUID]struct{}
}

func (h *Hosts) Add(hostID uuid.UUID) {
	h.Lock()
	defer h.Unlock()

	// Create the unhealthy hosts map if it doesn't exist.
	if h.hostIDs == nil {
		h.hostIDs = map[uuid.UUID]struct{}{}
	}

	h.hostIDs[hostID] = struct{}{}
}

func (h *Hosts) Remove(hostID uuid.UUID) {
	h.Lock()
	defer h.Unlock()

	delete(h.hostIDs, hostID)
}

// IsUnhealthy returns true if the host is unhealthy.
func (h *Hosts) IsUnhealthy(hostID uuid.UUID) bool {
	h.Lock()
	defer h.Unlock()

	// A nil map behaves like an empty map if read from, so this is okay.
	return h.hostIDs[hostID] != struct{}{}
}

// Length returns the length of unhealthy hosts.
func (h *Hosts) Length() int {
	h.Lock()
	defer h.Unlock()

	// A nil map behaves like an empty map if read from, so this is okay.
	return len(h.hostIDs)
}

// TODO add unhealthy host check-ins to watch for recovered hosts.
