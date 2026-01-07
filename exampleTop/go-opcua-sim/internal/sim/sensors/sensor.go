package sensors

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// Sensor represents a virtual sensor that generates data
type Sensor interface {
	// GetName returns the sensor name
	GetName() string

	// GetAddress returns the PLC address (e.g., "%DF100")
	GetAddress() string

	// Update generates the next sensor value based on elapsed time
	Update(deltaTime time.Duration) float64

	// Reset resets the sensor state
	Reset()

	// IsEnabled returns whether the sensor is active
	IsEnabled() bool
}

// BaseSensor provides common functionality for all sensors
type BaseSensor struct {
	Name              string
	Address           string
	Enabled           bool
	UpdateIntervalMs  int
	Description       string
	ElapsedTime       float64 // seconds
	mu                sync.RWMutex
}

// GetName returns the sensor name
func (b *BaseSensor) GetName() string {
	return b.Name
}

// GetAddress returns the sensor address
func (b *BaseSensor) GetAddress() string {
	return b.Address
}

// IsEnabled returns whether the sensor is enabled
func (b *BaseSensor) IsEnabled() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Enabled
}

// AddElapsedTime adds time to the elapsed counter (thread-safe)
func (b *BaseSensor) AddElapsedTime(deltaTime time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ElapsedTime += deltaTime.Seconds()
}

// GetElapsedTime returns the current elapsed time (thread-safe)
func (b *BaseSensor) GetElapsedTime() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.ElapsedTime
}

// Reset resets the base sensor state
func (b *BaseSensor) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ElapsedTime = 0
}

// gaussianNoise generates Gaussian noise with given mean and standard deviation
// Uses Box-Muller transform to generate normally distributed random numbers
func gaussianNoise(mean, stdDev float64) float64 {
	u1 := rand.Float64()
	u2 := rand.Float64()

	// Avoid log(0)
	if u1 < 1e-10 {
		u1 = 1e-10
	}

	// Box-Muller transform
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return mean + z0*stdDev
}
