package sensors

import (
	"time"
)

// RelayActuator simulates a relay/valve (digital output)
// This is a writable actuator that can be controlled via EPICS
type RelayActuator struct {
	BaseSensor
	DefaultState bool    // Default state when no command
	AutoToggle   bool    // Auto toggle for testing
	TogglePeriod float64 // Auto toggle period (seconds)
	CurrentState bool    // Current relay state
	LastToggle   float64 // Last toggle time
}

// NewRelayActuator creates a new relay actuator
func NewRelayActuator(name, address string, enabled bool, updateIntervalMs int,
	defaultState, autoToggle bool, togglePeriod float64, description string) *RelayActuator {
	return &RelayActuator{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		DefaultState: defaultState,
		AutoToggle:   autoToggle,
		TogglePeriod: togglePeriod,
		CurrentState: defaultState,
		LastToggle:   0,
	}
}

// Update returns the current state (can be modified externally via PLC write)
func (r *RelayActuator) Update(deltaTime time.Duration) float64 {
	if !r.IsEnabled() {
		return 0.0
	}

	r.AddElapsedTime(deltaTime)
	elapsed := r.GetElapsedTime()

	// Auto toggle mode (for testing)
	if r.AutoToggle && r.TogglePeriod > 0 {
		if elapsed-r.LastToggle >= r.TogglePeriod {
			r.CurrentState = !r.CurrentState
			r.LastToggle = elapsed
		}
	}

	if r.CurrentState {
		return 1.0
	}
	return 0.0
}

// SetState sets the relay state (called when EPICS writes to this address)
func (r *RelayActuator) SetState(state bool) {
	r.CurrentState = state
}

// Reset resets the actuator to default state
func (r *RelayActuator) Reset() {
	r.BaseSensor.Reset()
	r.CurrentState = r.DefaultState
	r.LastToggle = 0
}
