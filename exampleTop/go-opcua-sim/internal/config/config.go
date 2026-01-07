package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// SensorConfig represents the complete sensor configuration
type SensorConfig struct {
	Sensors []SensorDefinition `json:"sensors"`
}

// SensorDefinition defines a single sensor
type SensorDefinition struct {
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	Enabled          bool                   `json:"enabled"`
	Address          string                 `json:"address"`
	UpdateIntervalMs int                    `json:"updateIntervalMs"`
	Parameters       map[string]interface{} `json:"parameters"`
	Description      string                 `json:"description"`
}

// LoadConfig loads sensor configuration from a JSON file
func LoadConfig(filename string) (*SensorConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config SensorConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config JSON: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig validates the sensor configuration
func validateConfig(config *SensorConfig) error {
	if len(config.Sensors) == 0 {
		return fmt.Errorf("no sensors defined in configuration")
	}

	addressMap := make(map[string]bool)
	nameMap := make(map[string]bool)

	for i, sensor := range config.Sensors {
		// Check required fields
		if sensor.Name == "" {
			return fmt.Errorf("sensor at index %d has empty name", i)
		}
		if sensor.Type == "" {
			return fmt.Errorf("sensor '%s' has empty type", sensor.Name)
		}
		if sensor.Address == "" {
			return fmt.Errorf("sensor '%s' has empty address", sensor.Name)
		}
		if sensor.UpdateIntervalMs <= 0 {
			return fmt.Errorf("sensor '%s' has invalid updateIntervalMs: %d", sensor.Name, sensor.UpdateIntervalMs)
		}

		// Check for duplicate names
		if nameMap[sensor.Name] {
			return fmt.Errorf("duplicate sensor name: %s", sensor.Name)
		}
		nameMap[sensor.Name] = true

		// Check for duplicate addresses
		if addressMap[sensor.Address] {
			return fmt.Errorf("duplicate sensor address: %s (used by %s)", sensor.Address, sensor.Name)
		}
		addressMap[sensor.Address] = true

		// Validate address format
		if len(sensor.Address) < 4 || sensor.Address[0] != '%' {
			return fmt.Errorf("sensor '%s' has invalid address format: %s (expected format: %%DFxxx or %%DWxxx)", sensor.Name, sensor.Address)
		}
	}

	return nil
}

// GetFloat64Param safely retrieves a float64 parameter with default value
func GetFloat64Param(params map[string]interface{}, key string, defaultValue float64) float64 {
	if val, ok := params[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}

// GetBoolParam safely retrieves a bool parameter with default value
func GetBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}
