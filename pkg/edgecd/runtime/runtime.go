package runtime

import "sort"

// RuntimeState tracks state within a single reconciliation loop iteration.
// It accumulates services that need restarting and tracks whether a reboot is required.
type RuntimeState struct {
	ServicesToRestart map[string]bool // Set for deduplication
	RequireReboot     bool
}

// NewRuntimeState creates a new RuntimeState with empty state.
func NewRuntimeState() *RuntimeState {
	return &RuntimeState{
		ServicesToRestart: make(map[string]bool),
		RequireReboot:     false,
	}
}

// AddServiceRestart adds a service to the restart list.
// Duplicate services are automatically deduplicated via map.
func (rs *RuntimeState) AddServiceRestart(serviceName string) {
	rs.ServicesToRestart[serviceName] = true
}

// GetServicesToRestart returns a sorted, deduplicated list of services to restart.
func (rs *RuntimeState) GetServicesToRestart() []string {
	services := make([]string, 0, len(rs.ServicesToRestart))
	for service := range rs.ServicesToRestart {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

// Reset clears all state, preparing for the next reconciliation loop.
func (rs *RuntimeState) Reset() {
	rs.ServicesToRestart = make(map[string]bool)
	rs.RequireReboot = false
}
