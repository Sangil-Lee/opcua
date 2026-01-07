package sensors

import (
	"math"
	"math/rand"
	"time"
)

// DigitalSensor simulates digital/binary sensors (door, motion, limit switch, etc.)
type DigitalSensor struct {
	BaseSensor
	Pattern      string  // Pattern type: "toggle", "pulse", "random", "alarm"
	TogglePeriod float64 // Period for toggle pattern (seconds)
	PulseWidth   float64 // Pulse width (seconds)
	PulsePeriod  float64 // Period between pulses (seconds)
	RandomProb   float64 // Probability of being ON (0.0-1.0) for random pattern
	AlarmThreshold float64 // Threshold for alarm pattern
	CurrentState bool    // Current digital state
	LastToggle   float64 // Last toggle time
	rng          *rand.Rand
}

// NewDigitalSensor creates a new digital sensor
func NewDigitalSensor(name, address string, enabled bool, updateIntervalMs int,
	pattern string, togglePeriod, pulseWidth, pulsePeriod, randomProb float64,
	description string) *DigitalSensor {
	return &DigitalSensor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		Pattern:      pattern,
		TogglePeriod: togglePeriod,
		PulseWidth:   pulseWidth,
		PulsePeriod:  pulsePeriod,
		RandomProb:   randomProb,
		CurrentState: false,
		LastToggle:   0,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Update generates the next digital value based on pattern
func (d *DigitalSensor) Update(deltaTime time.Duration) float64 {
	if !d.IsEnabled() {
		return 0.0
	}

	d.AddElapsedTime(deltaTime)
	elapsed := d.GetElapsedTime()

	switch d.Pattern {
	case "toggle":
		// Toggle at regular intervals
		if elapsed-d.LastToggle >= d.TogglePeriod {
			d.CurrentState = !d.CurrentState
			d.LastToggle = elapsed
		}

	case "pulse":
		// Generate periodic pulses
		cycleTime := math.Mod(elapsed, d.PulsePeriod)
		d.CurrentState = cycleTime < d.PulseWidth

	case "random":
		// Random ON/OFF based on probability
		d.CurrentState = d.rng.Float64() < d.RandomProb

	case "alarm":
		// Alarm pattern: flashing when active
		// ON for 0.5s, OFF for 0.5s
		cycleTime := math.Mod(elapsed, 1.0)
		d.CurrentState = cycleTime < 0.5

	default:
		// Default: always OFF
		d.CurrentState = false
	}

	if d.CurrentState {
		return 1.0
	}
	return 0.0
}

// Reset resets the sensor to initial state
func (d *DigitalSensor) Reset() {
	d.BaseSensor.Reset()
	d.CurrentState = false
	d.LastToggle = 0
	d.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}
