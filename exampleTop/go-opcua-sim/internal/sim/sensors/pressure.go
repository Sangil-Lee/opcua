package sensors

import (
	"math"
	"math/rand"
	"time"
)

// PressureSensor simulates a pressure sensor with ramp up/hold/ramp down cycle
type PressureSensor struct {
	BaseSensor
	MinPressure  float64 // Minimum pressure value
	MaxPressure  float64 // Maximum pressure value
	RampUpTime   float64 // Time to ramp from min to max (seconds)
	HoldTime     float64 // Time to hold at max pressure (seconds)
	RampDownTime float64 // Time to ramp from max to min (seconds)
	NoiseStdDev  float64 // Standard deviation of Gaussian noise
	CyclePeriod  float64 // Total cycle period (calculated)
	rng          *rand.Rand
}

// NewPressureSensor creates a new pressure sensor
func NewPressureSensor(name, address string, enabled bool, updateIntervalMs int,
	minPressure, maxPressure, rampUpTime, holdTime, rampDownTime, noiseStdDev float64, description string) *PressureSensor {
	return &PressureSensor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		MinPressure:  minPressure,
		MaxPressure:  maxPressure,
		RampUpTime:   rampUpTime,
		HoldTime:     holdTime,
		RampDownTime: rampDownTime,
		NoiseStdDev:  noiseStdDev,
		CyclePeriod:  rampUpTime + holdTime + rampDownTime,
		rng:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Update generates the next pressure value based on the cycle phase
// Cycle: Ramp Up -> Hold -> Ramp Down -> (repeat)
func (p *PressureSensor) Update(deltaTime time.Duration) float64 {
	if !p.IsEnabled() {
		return p.MinPressure
	}

	p.AddElapsedTime(deltaTime)
	elapsed := p.GetElapsedTime()

	// Get position within current cycle
	cycleTime := math.Mod(elapsed, p.CyclePeriod)

	var pressure float64

	if cycleTime < p.RampUpTime {
		// Phase 1: Ramp Up
		progress := cycleTime / p.RampUpTime
		pressure = p.MinPressure + (p.MaxPressure-p.MinPressure)*progress

	} else if cycleTime < p.RampUpTime+p.HoldTime {
		// Phase 2: Hold at max
		pressure = p.MaxPressure

	} else {
		// Phase 3: Ramp Down
		rampDownStart := p.RampUpTime + p.HoldTime
		timeInRampDown := cycleTime - rampDownStart
		progress := timeInRampDown / p.RampDownTime
		pressure = p.MaxPressure - (p.MaxPressure-p.MinPressure)*progress
	}

	// Add Gaussian noise
	noise := p.generateGaussianNoise()
	pressure += noise

	// Clamp to valid range
	if pressure < p.MinPressure {
		pressure = p.MinPressure
	}
	if pressure > p.MaxPressure {
		pressure = p.MaxPressure
	}

	return pressure
}

// generateGaussianNoise generates Gaussian noise with mean=0 and stddev=NoiseStdDev
func (p *PressureSensor) generateGaussianNoise() float64 {
	// Box-Muller transform
	u1 := p.rng.Float64()
	u2 := p.rng.Float64()

	if u1 < 1e-10 {
		u1 = 1e-10
	}

	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return z0 * p.NoiseStdDev
}

// Reset resets the sensor to initial state
func (p *PressureSensor) Reset() {
	p.BaseSensor.Reset()
	p.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}
