package sensors

import (
	"math/rand"
	"time"
)

// RandomSensor generates random values with smooth transitions (random walk)
type RandomSensor struct {
	BaseSensor
	MinValue     float64 // Minimum value
	MaxValue     float64 // Maximum value
	ChangeRate   float64 // Maximum change per second
	CurrentValue float64 // Current value
	rng          *rand.Rand
}

// NewRandomSensor creates a new random sensor
func NewRandomSensor(name, address string, enabled bool, updateIntervalMs int,
	minValue, maxValue, changeRate float64, description string) *RandomSensor {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	initialValue := minValue + rng.Float64()*(maxValue-minValue)

	return &RandomSensor{
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
		ChangeRate:   changeRate,
		CurrentValue: initialValue,
		rng:          rng,
	}
}

// Update generates the next random value using random walk
func (r *RandomSensor) Update(deltaTime time.Duration) float64 {
	if !r.IsEnabled() {
		return r.CurrentValue
	}

	r.AddElapsedTime(deltaTime)

	// Random walk: add random change proportional to deltaTime
	maxChange := r.ChangeRate * deltaTime.Seconds()
	change := (r.rng.Float64()*2.0 - 1.0) * maxChange // Random value in [-maxChange, +maxChange]

	r.CurrentValue += change

	// Clamp to valid range
	if r.CurrentValue < r.MinValue {
		r.CurrentValue = r.MinValue
	}
	if r.CurrentValue > r.MaxValue {
		r.CurrentValue = r.MaxValue
	}

	return r.CurrentValue
}

// Reset resets the sensor to initial state
func (r *RandomSensor) Reset() {
	r.BaseSensor.Reset()
	r.CurrentValue = r.MinValue + r.rng.Float64()*(r.MaxValue-r.MinValue)
	r.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}
