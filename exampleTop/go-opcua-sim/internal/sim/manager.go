package sim

import (
	"fmt"
	"go-opcua-sim/internal/config"
	"go-opcua-sim/internal/plc"
	"go-opcua-sim/internal/sim/sensors"
	"log"
	"sync"
	"time"
)

// SensorManager manages all virtual sensors and updates the tag manager
type SensorManager struct {
	sensors      []sensors.Sensor
	tagManager   *plc.TagManager
	ticker       *time.Ticker
	stopChan     chan bool
	lastUpdate   time.Time
	mu           sync.RWMutex
	updateCount  uint64
}

// NewSensorManager creates a new sensor manager
func NewSensorManager(tagManager *plc.TagManager, cfg *config.SensorConfig) (*SensorManager, error) {
	manager := &SensorManager{
		sensors:    make([]sensors.Sensor, 0),
		tagManager: tagManager,
		stopChan:   make(chan bool),
		lastUpdate: time.Now(),
	}

	// Create sensors from configuration
	for _, def := range cfg.Sensors {
		sensor, err := CreateSensor(def)
		if err != nil {
			return nil, fmt.Errorf("failed to create sensor '%s': %w", def.Name, err)
		}
		manager.sensors = append(manager.sensors, sensor)
		log.Printf("Created sensor: %s (type=%s, address=%s)", def.Name, def.Type, def.Address)
	}

	if len(manager.sensors) == 0 {
		return nil, fmt.Errorf("no sensors created")
	}

	return manager, nil
}

// Start starts the sensor update loop
func (sm *SensorManager) Start(updateInterval time.Duration) {
	sm.ticker = time.NewTicker(updateInterval)
	sm.lastUpdate = time.Now()

	go func() {
		log.Printf("Sensor manager started with %d sensors (update interval: %v)", len(sm.sensors), updateInterval)

		for {
			select {
			case <-sm.ticker.C:
				sm.update()
			case <-sm.stopChan:
				log.Println("Sensor manager stopped")
				return
			}
		}
	}()
}

// Stop stops the sensor update loop
func (sm *SensorManager) Stop() {
	if sm.ticker != nil {
		sm.ticker.Stop()
	}
	close(sm.stopChan)
}

// update updates all sensors and writes values to tags
func (sm *SensorManager) update() {
	now := time.Now()
	deltaTime := now.Sub(sm.lastUpdate)
	sm.lastUpdate = now

	sm.mu.Lock()
	sm.updateCount++
	count := sm.updateCount
	sm.mu.Unlock()

	// Update all sensors in parallel
	var wg sync.WaitGroup
	for _, sensor := range sm.sensors {
		wg.Add(1)
		go func(s sensors.Sensor) {
			defer wg.Done()

			if !s.IsEnabled() {
				return
			}

			// Generate new value
			value := s.Update(deltaTime)

			// Write to tag manager
			if err := sm.tagManager.SetTagValue(s.GetName(), value); err != nil {
				log.Printf("Error writing sensor %s to tag: %v", s.GetName(), err)
			}
		}(sensor)
	}
	wg.Wait()

	// Log periodically (every 2 seconds = ~20 updates at 100ms interval)
	if count%20 == 0 {
		sm.logSensorValues()
	}
}

// logSensorValues logs current sensor values
func (sm *SensorManager) logSensorValues() {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	log.Printf("=== Sensor Update #%d ===", sm.updateCount)
	for _, sensor := range sm.sensors {
		if !sensor.IsEnabled() {
			continue
		}

		tag, err := sm.tagManager.GetTag(sensor.GetName())
		if err != nil {
			log.Printf("  %s: ERROR - %v", sensor.GetName(), err)
		} else {
			value, _ := tag.GetFloat64()
			log.Printf("  %s (%s): %.3f", sensor.GetName(), sensor.GetAddress(), value)
		}
	}
}

// GetSensorCount returns the number of managed sensors
func (sm *SensorManager) GetSensorCount() int {
	return len(sm.sensors)
}

// GetUpdateCount returns the total number of updates performed
func (sm *SensorManager) GetUpdateCount() uint64 {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.updateCount
}

// GetAllSensors returns all sensors
func (sm *SensorManager) GetAllSensors() []sensors.Sensor {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.sensors
}

// GetSensor returns a sensor by name
func (sm *SensorManager) GetSensor(name string) sensors.Sensor {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	for _, sensor := range sm.sensors {
		if sensor.GetName() == name {
			return sensor
		}
	}
	return nil
}
