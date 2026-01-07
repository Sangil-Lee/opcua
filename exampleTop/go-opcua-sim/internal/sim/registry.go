package sim

import (
	"fmt"
	"go-opcua-sim/internal/config"
	"go-opcua-sim/internal/sim/sensors"
)

// SensorFactory is a function that creates a sensor from configuration
type SensorFactory func(def config.SensorDefinition) (sensors.Sensor, error)

// sensorRegistry holds registered sensor factories
var sensorRegistry = make(map[string]SensorFactory)

// RegisterSensor registers a sensor factory for a given type
func RegisterSensor(sensorType string, factory SensorFactory) {
	sensorRegistry[sensorType] = factory
}

// CreateSensor creates a sensor instance from configuration
func CreateSensor(def config.SensorDefinition) (sensors.Sensor, error) {
	factory, ok := sensorRegistry[def.Type]
	if !ok {
		return nil, fmt.Errorf("unknown sensor type: %s", def.Type)
	}

	sensor, err := factory(def)
	if err != nil {
		return nil, fmt.Errorf("failed to create sensor '%s': %w", def.Name, err)
	}

	return sensor, nil
}

// init registers built-in sensor types
func init() {
	// Register Temperature sensor
	RegisterSensor("temperature", func(def config.SensorDefinition) (sensors.Sensor, error) {
		baseTemp := config.GetFloat64Param(def.Parameters, "baseTemp", 25.0)
		amplitude := config.GetFloat64Param(def.Parameters, "amplitude", 10.0)
		period := config.GetFloat64Param(def.Parameters, "period", 30.0)
		noiseStdDev := config.GetFloat64Param(def.Parameters, "noiseStdDev", 0.5)
		minValue := config.GetFloat64Param(def.Parameters, "minValue", 0.0)
		maxValue := config.GetFloat64Param(def.Parameters, "maxValue", 100.0)

		return sensors.NewTemperatureSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			baseTemp, amplitude, period, noiseStdDev, minValue, maxValue,
			def.Description,
		), nil
	})

	// Register Pressure sensor
	RegisterSensor("pressure", func(def config.SensorDefinition) (sensors.Sensor, error) {
		minPressure := config.GetFloat64Param(def.Parameters, "minPressure", 0.0)
		maxPressure := config.GetFloat64Param(def.Parameters, "maxPressure", 10.0)
		rampUpTime := config.GetFloat64Param(def.Parameters, "rampUpTime", 20.0)
		holdTime := config.GetFloat64Param(def.Parameters, "holdTime", 10.0)
		rampDownTime := config.GetFloat64Param(def.Parameters, "rampDownTime", 15.0)
		noiseStdDev := config.GetFloat64Param(def.Parameters, "noiseStdDev", 0.1)

		return sensors.NewPressureSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			minPressure, maxPressure, rampUpTime, holdTime, rampDownTime, noiseStdDev,
			def.Description,
		), nil
	})

	// Register Sine sensor
	RegisterSensor("sine", func(def config.SensorDefinition) (sensors.Sensor, error) {
		offset := config.GetFloat64Param(def.Parameters, "offset", 0.0)
		amplitude := config.GetFloat64Param(def.Parameters, "amplitude", 1.0)
		frequency := config.GetFloat64Param(def.Parameters, "frequency", 1.0)
		phase := config.GetFloat64Param(def.Parameters, "phase", 0.0)

		return sensors.NewSineSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			offset, amplitude, frequency, phase,
			def.Description,
		), nil
	})

	// Register Random sensor
	RegisterSensor("random", func(def config.SensorDefinition) (sensors.Sensor, error) {
		minValue := config.GetFloat64Param(def.Parameters, "minValue", 0.0)
		maxValue := config.GetFloat64Param(def.Parameters, "maxValue", 100.0)
		changeRate := config.GetFloat64Param(def.Parameters, "changeRate", 1.0)

		return sensors.NewRandomSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			minValue, maxValue, changeRate,
			def.Description,
		), nil
	})

	// Register Digital sensor
	RegisterSensor("digital", func(def config.SensorDefinition) (sensors.Sensor, error) {
		pattern := "toggle"
		if p, ok := def.Parameters["pattern"].(string); ok {
			pattern = p
		}
		togglePeriod := config.GetFloat64Param(def.Parameters, "togglePeriod", 5.0)
		pulseWidth := config.GetFloat64Param(def.Parameters, "pulseWidth", 1.0)
		pulsePeriod := config.GetFloat64Param(def.Parameters, "pulsePeriod", 5.0)
		randomProb := config.GetFloat64Param(def.Parameters, "randomProb", 0.5)

		return sensors.NewDigitalSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			pattern, togglePeriod, pulseWidth, pulsePeriod, randomProb,
			def.Description,
		), nil
	})

	// Register Vibration sensor
	RegisterSensor("vibration", func(def config.SensorDefinition) (sensors.Sensor, error) {
		baseLevel := config.GetFloat64Param(def.Parameters, "baseLevel", 2.0)
		amplitude := config.GetFloat64Param(def.Parameters, "amplitude", 1.0)
		frequency := config.GetFloat64Param(def.Parameters, "frequency", 50.0)
		harmonics := int(config.GetFloat64Param(def.Parameters, "harmonics", 3))
		spikeProb := config.GetFloat64Param(def.Parameters, "spikeProb", 0.01)
		spikeAmp := config.GetFloat64Param(def.Parameters, "spikeAmp", 5.0)
		noiseStdDev := config.GetFloat64Param(def.Parameters, "noiseStdDev", 0.2)
		minValue := config.GetFloat64Param(def.Parameters, "minValue", 0.0)
		maxValue := config.GetFloat64Param(def.Parameters, "maxValue", 50.0)

		return sensors.NewVibrationSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			baseLevel, amplitude, frequency, harmonics, spikeProb, spikeAmp, noiseStdDev, minValue, maxValue,
			def.Description,
		), nil
	})

	// Register Noise sensor
	RegisterSensor("noise", func(def config.SensorDefinition) (sensors.Sensor, error) {
		ambientLevel := config.GetFloat64Param(def.Parameters, "ambientLevel", 55.0)
		peakLevel := config.GetFloat64Param(def.Parameters, "peakLevel", 85.0)
		cyclePeriod := config.GetFloat64Param(def.Parameters, "cyclePeriod", 30.0)
		dutyCycle := config.GetFloat64Param(def.Parameters, "dutyCycle", 0.6)
		noiseStdDev := config.GetFloat64Param(def.Parameters, "noiseStdDev", 2.0)
		spikeProb := config.GetFloat64Param(def.Parameters, "spikeProb", 0.02)
		spikeLevel := config.GetFloat64Param(def.Parameters, "spikeLevel", 95.0)
		minValue := config.GetFloat64Param(def.Parameters, "minValue", 30.0)
		maxValue := config.GetFloat64Param(def.Parameters, "maxValue", 120.0)

		return sensors.NewNoiseSensor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			ambientLevel, peakLevel, cyclePeriod, dutyCycle, noiseStdDev, spikeProb, spikeLevel, minValue, maxValue,
			def.Description,
		), nil
	})

	// Register Relay actuator
	RegisterSensor("relay", func(def config.SensorDefinition) (sensors.Sensor, error) {
		defaultState := config.GetBoolParam(def.Parameters, "defaultState", false)
		autoToggle := config.GetBoolParam(def.Parameters, "autoToggle", false)
		togglePeriod := config.GetFloat64Param(def.Parameters, "togglePeriod", 10.0)

		return sensors.NewRelayActuator(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			defaultState, autoToggle, togglePeriod,
			def.Description,
		), nil
	})

	// Register Integer actuator
	RegisterSensor("integer", func(def config.SensorDefinition) (sensors.Sensor, error) {
		minValue := int(config.GetFloat64Param(def.Parameters, "minValue", 0))
		maxValue := int(config.GetFloat64Param(def.Parameters, "maxValue", 100))
		defaultValue := int(config.GetFloat64Param(def.Parameters, "defaultValue", 0))
		autoMode := config.GetBoolParam(def.Parameters, "autoMode", false)
		autoPattern := "ramp"
		if p, ok := def.Parameters["autoPattern"].(string); ok {
			autoPattern = p
		}
		rampRate := config.GetFloat64Param(def.Parameters, "rampRate", 1.0)
		stepPeriod := config.GetFloat64Param(def.Parameters, "stepPeriod", 10.0)

		return sensors.NewIntegerActuator(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			minValue, maxValue, defaultValue, autoMode, autoPattern, rampRate, stepPeriod,
			def.Description,
		), nil
	})

	// Register Step Motor
	RegisterSensor("stepmotor", func(def config.SensorDefinition) (sensors.Sensor, error) {
		maxSpeed := config.GetFloat64Param(def.Parameters, "maxSpeed", 1000.0)
		acceleration := config.GetFloat64Param(def.Parameters, "acceleration", 5000.0)
		stepsPerRev := int(config.GetFloat64Param(def.Parameters, "stepsPerRev", 200))
		autoMode := config.GetBoolParam(def.Parameters, "autoMode", true)
		autoPattern := "oscillate"
		if p, ok := def.Parameters["autoPattern"].(string); ok {
			autoPattern = p
		}
		autoPeriod := config.GetFloat64Param(def.Parameters, "autoPeriod", 20.0)

		return sensors.NewStepMotor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			maxSpeed, acceleration, stepsPerRev, autoMode, autoPattern, autoPeriod,
			def.Description,
		), nil
	})

	// Register Servo Motor
	RegisterSensor("servomotor", func(def config.SensorDefinition) (sensors.Sensor, error) {
		maxVelocity := config.GetFloat64Param(def.Parameters, "maxVelocity", 3000.0)
		maxTorque := config.GetFloat64Param(def.Parameters, "maxTorque", 10.0)
		inertia := config.GetFloat64Param(def.Parameters, "inertia", 0.001)
		damping := config.GetFloat64Param(def.Parameters, "damping", 0.01)
		outputMode := "velocity"
		if p, ok := def.Parameters["outputMode"].(string); ok {
			outputMode = p
		}
		autoMode := config.GetBoolParam(def.Parameters, "autoMode", true)
		autoPattern := "sine"
		if p, ok := def.Parameters["autoPattern"].(string); ok {
			autoPattern = p
		}
		autoPeriod := config.GetFloat64Param(def.Parameters, "autoPeriod", 30.0)
		kp := config.GetFloat64Param(def.Parameters, "kp", 0.5)
		ki := config.GetFloat64Param(def.Parameters, "ki", 0.1)
		kd := config.GetFloat64Param(def.Parameters, "kd", 0.01)

		return sensors.NewServoMotor(
			def.Name, def.Address, def.Enabled, def.UpdateIntervalMs,
			maxVelocity, maxTorque, inertia, damping,
			outputMode, autoMode, autoPattern, autoPeriod,
			kp, ki, kd,
			def.Description,
		), nil
	})
}
