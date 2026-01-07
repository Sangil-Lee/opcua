package sensors

import (
	"math"
	"time"
)

// IntegerActuator simulates integer-based actuators (motor speed, heater power, etc.)
type IntegerActuator struct {
	BaseSensor
	MinValue     int     // Minimum value
	MaxValue     int     // Maximum value
	DefaultValue int     // Default value
	AutoMode     bool    // Auto mode for testing
	AutoPattern  string  // Pattern: "ramp", "sine", "step"
	RampRate     float64 // Ramp rate (units per second)
	StepPeriod   float64 // Step period (seconds)
	CurrentValue int     // Current value
	TargetValue  int     // Target value (for ramping)
}

// NewIntegerActuator creates a new integer actuator
func NewIntegerActuator(name, address string, enabled bool, updateIntervalMs int,
	minValue, maxValue, defaultValue int, autoMode bool, autoPattern string,
	rampRate, stepPeriod float64, description string) *IntegerActuator {
	return &IntegerActuator{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		MinValue:     minValue,
		MaxValue:     maxValue,
		DefaultValue: defaultValue,
		AutoMode:     autoMode,
		AutoPattern:  autoPattern,
		RampRate:     rampRate,
		StepPeriod:   stepPeriod,
		CurrentValue: defaultValue,
		TargetValue:  defaultValue,
	}
}

// Update returns the current integer value
func (i *IntegerActuator) Update(deltaTime time.Duration) float64 {
	if !i.IsEnabled() {
		return float64(i.DefaultValue)
	}

	i.AddElapsedTime(deltaTime)
	elapsed := i.GetElapsedTime()

	if i.AutoMode {
		switch i.AutoPattern {
		case "ramp":
			// Ramp up and down between min and max
			rampPeriod := float64(i.MaxValue-i.MinValue) / i.RampRate
			totalPeriod := rampPeriod * 2 // Up and down
			cycleTime := math.Mod(elapsed, totalPeriod)

			if cycleTime < rampPeriod {
				// Ramp up
				progress := cycleTime / rampPeriod
				i.CurrentValue = i.MinValue + int(float64(i.MaxValue-i.MinValue)*progress)
			} else {
				// Ramp down
				progress := (cycleTime - rampPeriod) / rampPeriod
				i.CurrentValue = i.MaxValue - int(float64(i.MaxValue-i.MinValue)*progress)
			}

		case "sine":
			// Sine wave between min and max
			period := i.StepPeriod
			midValue := (i.MaxValue + i.MinValue) / 2
			amplitude := (i.MaxValue - i.MinValue) / 2
			value := float64(midValue) + float64(amplitude)*math.Sin(2.0*math.Pi*elapsed/period)
			i.CurrentValue = int(value)

		case "step":
			// Step pattern
			cycleTime := math.Mod(elapsed, i.StepPeriod*2)
			if cycleTime < i.StepPeriod {
				i.CurrentValue = i.MaxValue
			} else {
				i.CurrentValue = i.MinValue
			}
		}
	}

	// Clamp to valid range
	if i.CurrentValue < i.MinValue {
		i.CurrentValue = i.MinValue
	}
	if i.CurrentValue > i.MaxValue {
		i.CurrentValue = i.MaxValue
	}

	return float64(i.CurrentValue)
}

// SetValue sets the actuator value (called when EPICS writes to this address)
func (i *IntegerActuator) SetValue(value int) {
	if value < i.MinValue {
		value = i.MinValue
	}
	if value > i.MaxValue {
		value = i.MaxValue
	}
	i.CurrentValue = value
	i.TargetValue = value
}

// Reset resets the actuator to default value
func (i *IntegerActuator) Reset() {
	i.BaseSensor.Reset()
	i.CurrentValue = i.DefaultValue
	i.TargetValue = i.DefaultValue
}
