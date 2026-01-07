package sensors

import (
	"math"
	"math/rand"
	"time"
)

// VibrationSensor simulates vibration sensors (accelerometers) measuring machinery vibration
// Output is in mm/s (velocity RMS) which is common for industrial vibration monitoring
type VibrationSensor struct {
	BaseSensor
	BaseLevel     float64 // Base vibration level (mm/s RMS)
	Amplitude     float64 // Amplitude of primary oscillation
	Frequency     float64 // Primary frequency (Hz)
	Harmonics     int     // Number of harmonics to include
	SpikeProb     float64 // Probability of vibration spike per update (0.0-1.0)
	SpikeAmp      float64 // Amplitude of vibration spikes
	NoiseStdDev   float64 // Standard deviation of Gaussian noise
	MinValue      float64 // Minimum vibration value
	MaxValue      float64 // Maximum vibration value
	lastSpikeTime float64 // Track last spike time
	spikeDecay    float64 // Current spike decay value
}

// NewVibrationSensor creates a new vibration sensor
func NewVibrationSensor(name, address string, enabled bool, updateIntervalMs int,
	baseLevel, amplitude, frequency float64, harmonics int, spikeProb, spikeAmp,
	noiseStdDev, minValue, maxValue float64, description string) *VibrationSensor {
	return &VibrationSensor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		BaseLevel:     baseLevel,
		Amplitude:     amplitude,
		Frequency:     frequency,
		Harmonics:     harmonics,
		SpikeProb:     spikeProb,
		SpikeAmp:      spikeAmp,
		NoiseStdDev:   noiseStdDev,
		MinValue:      minValue,
		MaxValue:      maxValue,
		lastSpikeTime: 0,
		spikeDecay:    0,
	}
}

// Update calculates the current vibration level
func (v *VibrationSensor) Update(deltaTime time.Duration) float64 {
	if !v.IsEnabled() {
		return 0.0
	}

	v.AddElapsedTime(deltaTime)
	elapsed := v.GetElapsedTime()

	// Calculate base vibration with multiple frequency components (harmonics)
	vibration := v.BaseLevel

	// Add fundamental frequency and harmonics
	for i := 1; i <= v.Harmonics; i++ {
		harmonicFreq := v.Frequency * float64(i)
		harmonicAmp := v.Amplitude / float64(i) // Harmonics decrease in amplitude
		vibration += harmonicAmp * math.Sin(2.0*math.Pi*harmonicFreq*elapsed)
	}

	// Add random vibration spikes (simulating impacts, bearing defects, etc.)
	if rand.Float64() < v.SpikeProb {
		v.lastSpikeTime = elapsed
		v.spikeDecay = v.SpikeAmp
	}

	// Apply exponential decay to spike
	if v.spikeDecay > 0.01 {
		timeSinceSpike := elapsed - v.lastSpikeTime
		v.spikeDecay = v.SpikeAmp * math.Exp(-timeSinceSpike*5.0) // Fast decay
		vibration += v.spikeDecay
	}

	// Add Gaussian noise (random vibration components)
	if v.NoiseStdDev > 0 {
		vibration += gaussianNoise(0, v.NoiseStdDev)
	}

	// Ensure non-negative (vibration is always positive as RMS value)
	if vibration < 0 {
		vibration = 0
	}

	// Clamp to min/max range
	if vibration < v.MinValue {
		vibration = v.MinValue
	}
	if vibration > v.MaxValue {
		vibration = v.MaxValue
	}

	return vibration
}

// Reset resets the sensor to initial state
func (v *VibrationSensor) Reset() {
	v.BaseSensor.Reset()
	v.lastSpikeTime = 0
	v.spikeDecay = 0
}
