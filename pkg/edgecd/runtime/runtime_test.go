package runtime

import (
	"reflect"
	"testing"
)

func TestNewRuntimeState(t *testing.T) {
	state := NewRuntimeState()

	if state == nil {
		t.Fatal("NewRuntimeState() returned nil")
	}

	if state.ServicesToRestart == nil {
		t.Error("ServicesToRestart map not initialized")
	}

	if len(state.ServicesToRestart) != 0 {
		t.Errorf("Expected empty ServicesToRestart, got %d entries", len(state.ServicesToRestart))
	}

	if state.RequireReboot {
		t.Error("RequireReboot should be false initially")
	}
}

func TestAddServiceRestart(t *testing.T) {
	state := NewRuntimeState()

	// Add first service
	state.AddServiceRestart("nginx")
	if !state.ServicesToRestart["nginx"] {
		t.Error("nginx not added to ServicesToRestart")
	}

	// Add second service
	state.AddServiceRestart("redis")
	if !state.ServicesToRestart["redis"] {
		t.Error("redis not added to ServicesToRestart")
	}

	// Add duplicate service (should be idempotent)
	state.AddServiceRestart("nginx")
	if len(state.ServicesToRestart) != 2 {
		t.Errorf("Expected 2 services after adding duplicate, got %d", len(state.ServicesToRestart))
	}
}

func TestGetServicesToRestart(t *testing.T) {
	tests := []struct {
		name     string
		services []string
		want     []string
	}{
		{
			name:     "empty state",
			services: []string{},
			want:     []string{},
		},
		{
			name:     "single service",
			services: []string{"nginx"},
			want:     []string{"nginx"},
		},
		{
			name:     "multiple services sorted",
			services: []string{"zebra", "alpha", "beta"},
			want:     []string{"alpha", "beta", "zebra"},
		},
		{
			name:     "with duplicates",
			services: []string{"nginx", "redis", "nginx", "postgres"},
			want:     []string{"nginx", "postgres", "redis"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := NewRuntimeState()

			// Add services
			for _, svc := range tt.services {
				state.AddServiceRestart(svc)
			}

			// Get sorted list
			got := state.GetServicesToRestart()

			// Compare
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetServicesToRestart() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReset(t *testing.T) {
	state := NewRuntimeState()

	// Add some state
	state.AddServiceRestart("nginx")
	state.AddServiceRestart("redis")
	state.RequireReboot = true

	// Verify state is populated
	if len(state.ServicesToRestart) != 2 {
		t.Error("State not populated correctly before reset")
	}
	if !state.RequireReboot {
		t.Error("RequireReboot not set before reset")
	}

	// Reset
	state.Reset()

	// Verify state is cleared
	if len(state.ServicesToRestart) != 0 {
		t.Errorf("ServicesToRestart not cleared, got %d entries", len(state.ServicesToRestart))
	}

	if state.RequireReboot {
		t.Error("RequireReboot not cleared after reset")
	}
}

func TestRequireReboot(t *testing.T) {
	state := NewRuntimeState()

	// Initially false
	if state.RequireReboot {
		t.Error("RequireReboot should be false initially")
	}

	// Set to true
	state.RequireReboot = true
	if !state.RequireReboot {
		t.Error("RequireReboot should be true after setting")
	}

	// Reset clears it
	state.Reset()
	if state.RequireReboot {
		t.Error("RequireReboot should be false after reset")
	}
}
