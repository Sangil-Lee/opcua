package sensors

import (
	"math"
	"time"
)

// SineSensor generates a simple sine wave
type SineSensor struct {
	BaseSensor
	Offset    float64 // DC offset
	Amplitude float64 // Amplitude of sine wave
	Frequency float64 // Frequency in Hz
	Phase     float64 // Phase offset in radians
}

// NewSineSensor creates a new sine wave sensor
func NewSineSensor(name, address string, enabled bool, updateIntervalMs int,
	offset, amplitude, frequency, phase float64, description string) *SineSensor {
	return &SineSensor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		Offset:    offset,
		Amplitude: amplitude,
		Frequency: frequency,
		Phase:     phase,
	}
}

// Update generates the next sine wave value
// Value = Offset + Amplitude * sin(2π * Frequency * t + Phase)
func (s *SineSensor) Update(deltaTime time.Duration) float64 {
	if !s.IsEnabled() {
		return s.Offset
	}

	s.AddElapsedTime(deltaTime)
	elapsed := s.GetElapsedTime()

	value := s.Offset + s.Amplitude*math.Sin(2.0*math.Pi*s.Frequency*elapsed+s.Phase)
	return value
}

// Reset resets the sensor to initial state
func (s *SineSensor) Reset() {
	s.BaseSensor.Reset()
}
