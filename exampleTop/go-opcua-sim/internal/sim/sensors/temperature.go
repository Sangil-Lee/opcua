package sensors

import (
	"math"
	"math/rand"
	"time"
)

// TemperatureSensor simulates a temperature sensor with sinusoidal variation and Gaussian noise
type TemperatureSensor struct {
	BaseSensor
	BaseTemp     float64 // Base temperature (offset)
	Amplitude    float64 // Temperature variation amplitude
	Period       float64 // Period in seconds for one complete cycle
	NoiseStdDev  float64 // Standard deviation of Gaussian noise
	MinValue     float64 // Minimum allowed temperature
	MaxValue     float64 // Maximum allowed temperature
	rng          *rand.Rand
}

// NewTemperatureSensor creates a new temperature sensor
func NewTemperatureSensor(name, address string, enabled bool, updateIntervalMs int,
	baseTemp, amplitude, period, noiseStdDev, minValue, maxValue float64, description string) *TemperatureSensor {
	return &TemperatureSensor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		BaseTemp:    baseTemp,
		Amplitude:   amplitude,
		Period:      period,
		NoiseStdDev: noiseStdDev,
		MinValue:    minValue,
		MaxValue:    maxValue,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Update generates the next temperature value
// Temperature = BaseTemp + Amplitude * sin(2π * t / Period) + GaussianNoise
func (t *TemperatureSensor) Update(deltaTime time.Duration) float64 {
	if !t.IsEnabled() {
		return t.BaseTemp
	}

	t.AddElapsedTime(deltaTime)
	elapsed := t.GetElapsedTime()

	// Sinusoidal component
	angularFrequency := 2.0 * math.Pi / t.Period
	sineValue := math.Sin(angularFrequency * elapsed)
	temperature := t.BaseTemp + t.Amplitude*sineValue

	// Add Gaussian noise using Box-Muller transform
	noise := t.generateGaussianNoise()
	temperature += noise

	// Clamp to valid range
	if temperature < t.MinValue {
		temperature = t.MinValue
	}
	if temperature > t.MaxValue {
		temperature = t.MaxValue
	}

	return temperature
}

// generateGaussianNoise generates Gaussian noise with mean=0 and stddev=NoiseStdDev
func (t *TemperatureSensor) generateGaussianNoise() float64 {
	// Box-Muller transform for Gaussian distribution
	u1 := t.rng.Float64()
	u2 := t.rng.Float64()

	// Avoid log(0)
	if u1 < 1e-10 {
		u1 = 1e-10
	}

	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return z0 * t.NoiseStdDev
}

// Reset resets the sensor to initial state
func (t *TemperatureSensor) Reset() {
	t.BaseSensor.Reset()
	t.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
}
