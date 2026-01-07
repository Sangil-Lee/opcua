package sensors

import (
	"math"
	"math/rand"
	"time"
)

// NoiseSensor simulates sound level sensors (microphones) measuring ambient noise
// Output is in dB (A-weighted decibels) which is standard for industrial noise monitoring
type NoiseSensor struct {
	BaseSensor
	AmbientLevel  float64 // Ambient/base noise level (dB)
	PeakLevel     float64 // Peak noise level during machinery operation (dB)
	CyclePeriod   float64 // Period of machinery cycle (seconds)
	DutyCycle     float64 // Fraction of cycle with high noise (0.0-1.0)
	NoiseStdDev   float64 // Standard deviation of random noise variation
	SpikeProb     float64 // Probability of noise spike per update (0.0-1.0)
	SpikeLevel    float64 // Level of noise spikes (dB)
	MinValue      float64 // Minimum noise value (dB)
	MaxValue      float64 // Maximum noise value (dB)
	lastSpikeTime float64 // Track last spike time
	spikeActive   bool    // Whether spike is currently active
}

// NewNoiseSensor creates a new noise sensor
func NewNoiseSensor(name, address string, enabled bool, updateIntervalMs int,
	ambientLevel, peakLevel, cyclePeriod, dutyCycle, noiseStdDev, spikeProb,
	spikeLevel, minValue, maxValue float64, description string) *NoiseSensor {
	return &NoiseSensor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		AmbientLevel:  ambientLevel,
		PeakLevel:     peakLevel,
		CyclePeriod:   cyclePeriod,
		DutyCycle:     dutyCycle,
		NoiseStdDev:   noiseStdDev,
		SpikeProb:     spikeProb,
		SpikeLevel:    spikeLevel,
		MinValue:      minValue,
		MaxValue:      maxValue,
		lastSpikeTime: 0,
		spikeActive:   false,
	}
}

// Update calculates the current noise level
func (n *NoiseSensor) Update(deltaTime time.Duration) float64 {
	if !n.IsEnabled() {
		return n.AmbientLevel
	}

	n.AddElapsedTime(deltaTime)
	elapsed := n.GetElapsedTime()

	// Calculate base noise level based on machinery cycle
	var baseNoise float64
	cyclePosition := math.Mod(elapsed, n.CyclePeriod) / n.CyclePeriod

	if cyclePosition < n.DutyCycle {
		// High noise period (machinery operating)
		// Smooth transition using sine interpolation
		progress := cyclePosition / n.DutyCycle
		transitionFactor := (1.0 - math.Cos(progress*math.Pi)) / 2.0
		baseNoise = n.AmbientLevel + (n.PeakLevel-n.AmbientLevel)*transitionFactor
	} else {
		// Low noise period (machinery idle)
		progress := (cyclePosition - n.DutyCycle) / (1.0 - n.DutyCycle)
		transitionFactor := (1.0 - math.Cos(progress*math.Pi)) / 2.0
		baseNoise = n.PeakLevel - (n.PeakLevel-n.AmbientLevel)*transitionFactor
	}

	// Add random noise spikes (impacts, drops, alarms, etc.)
	if rand.Float64() < n.SpikeProb {
		n.lastSpikeTime = elapsed
		n.spikeActive = true
	}

	// Apply spike with exponential decay
	if n.spikeActive {
		timeSinceSpike := elapsed - n.lastSpikeTime
		if timeSinceSpike < 0.3 { // Spike lasts 0.3 seconds
			spikeDecay := math.Exp(-timeSinceSpike * 10.0)
			// Convert to dB scale (logarithmic addition)
			spikeContribution := n.SpikeLevel * spikeDecay
			baseNoise = 10.0 * math.Log10(math.Pow(10.0, baseNoise/10.0)+math.Pow(10.0, spikeContribution/10.0))
		} else {
			n.spikeActive = false
		}
	}

	// Add Gaussian noise for natural variation
	if n.NoiseStdDev > 0 {
		baseNoise += gaussianNoise(0, n.NoiseStdDev)
	}

	// Clamp to min/max range
	if baseNoise < n.MinValue {
		baseNoise = n.MinValue
	}
	if baseNoise > n.MaxValue {
		baseNoise = n.MaxValue
	}

	return baseNoise
}

// Reset resets the sensor to initial state
func (n *NoiseSensor) Reset() {
	n.BaseSensor.Reset()
	n.lastSpikeTime = 0
	n.spikeActive = false
}
