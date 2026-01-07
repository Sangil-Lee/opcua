package sensors

import (
	"math"
	"time"
)

// StepMotor simulates a step motor with position control
// Outputs current position in steps
type StepMotor struct {
	BaseSensor
	MaxSpeed        float64 // Maximum speed (steps/second)
	Acceleration    float64 // Acceleration (steps/second^2)
	StepsPerRev     int     // Steps per revolution
	CurrentPosition float64 // Current position (steps)
	CurrentVelocity float64 // Current velocity (steps/second)
	TargetPosition  float64 // Target position (steps)
	AutoMode        bool    // Auto mode for demo
	AutoPattern     string  // Pattern: "oscillate", "rotate", "step"
	AutoPeriod      float64 // Period for auto pattern (seconds)
	Enabled         bool    // Motor enabled state
}

// NewStepMotor creates a new step motor
func NewStepMotor(name, address string, enabled bool, updateIntervalMs int,
	maxSpeed, acceleration float64, stepsPerRev int, autoMode bool, autoPattern string,
	autoPeriod float64, description string) *StepMotor {
	return &StepMotor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		MaxSpeed:        maxSpeed,
		Acceleration:    acceleration,
		StepsPerRev:     stepsPerRev,
		CurrentPosition: 0,
		CurrentVelocity: 0,
		TargetPosition:  0,
		AutoMode:        autoMode,
		AutoPattern:     autoPattern,
		AutoPeriod:      autoPeriod,
	}
}

// Update calculates the current motor position
func (s *StepMotor) Update(deltaTime time.Duration) float64 {
	if !s.IsEnabled() {
		// Motor disabled, maintain position but velocity goes to zero
		s.CurrentVelocity = 0
		return s.CurrentPosition
	}

	s.AddElapsedTime(deltaTime)
	elapsed := s.GetElapsedTime()
	dt := deltaTime.Seconds()

	// Auto mode: automatically change target position
	if s.AutoMode {
		switch s.AutoPattern {
		case "oscillate":
			// Oscillate between 0 and stepsPerRev
			progress := math.Mod(elapsed, s.AutoPeriod) / s.AutoPeriod
			if progress < 0.5 {
				s.TargetPosition = float64(s.StepsPerRev) * (progress * 2.0)
			} else {
				s.TargetPosition = float64(s.StepsPerRev) * (2.0 - progress*2.0)
			}

		case "rotate":
			// Continuous rotation
			rotationsPerSec := 1.0 / s.AutoPeriod
			s.TargetPosition = rotationsPerSec * float64(s.StepsPerRev) * elapsed

		case "step":
			// Step pattern: move to different positions periodically
			stepIndex := int(elapsed/s.AutoPeriod) % 4
			positions := []float64{0, float64(s.StepsPerRev) / 4, float64(s.StepsPerRev) / 2, float64(s.StepsPerRev) * 3 / 4}
			s.TargetPosition = positions[stepIndex]
		}
	}

	// Calculate position error
	posError := s.TargetPosition - s.CurrentPosition

	// Calculate desired velocity using trapezoidal profile
	var desiredVelocity float64
	if math.Abs(posError) < 0.01 {
		// At target, stop
		desiredVelocity = 0
	} else {
		// Move towards target
		direction := 1.0
		if posError < 0 {
			direction = -1.0
		}

		// Calculate deceleration distance
		decelDist := (s.CurrentVelocity * s.CurrentVelocity) / (2.0 * s.Acceleration)

		if math.Abs(posError) <= decelDist {
			// Decelerate
			desiredVelocity = direction * math.Sqrt(2.0*s.Acceleration*math.Abs(posError))
		} else {
			// Accelerate or maintain max speed
			desiredVelocity = direction * s.MaxSpeed
		}
	}

	// Update velocity with acceleration limit
	velError := desiredVelocity - s.CurrentVelocity
	maxVelChange := s.Acceleration * dt

	if math.Abs(velError) <= maxVelChange {
		s.CurrentVelocity = desiredVelocity
	} else {
		if velError > 0 {
			s.CurrentVelocity += maxVelChange
		} else {
			s.CurrentVelocity -= maxVelChange
		}
	}

	// Clamp velocity to max speed
	if s.CurrentVelocity > s.MaxSpeed {
		s.CurrentVelocity = s.MaxSpeed
	} else if s.CurrentVelocity < -s.MaxSpeed {
		s.CurrentVelocity = -s.MaxSpeed
	}

	// Update position
	s.CurrentPosition += s.CurrentVelocity * dt

	// Wrap position for continuous rotation mode
	if s.AutoPattern == "rotate" {
		s.CurrentPosition = math.Mod(s.CurrentPosition, float64(s.StepsPerRev))
		if s.CurrentPosition < 0 {
			s.CurrentPosition += float64(s.StepsPerRev)
		}
	}

	return s.CurrentPosition
}

// GetVelocity returns the current velocity (for monitoring)
func (s *StepMotor) GetVelocity() float64 {
	return s.CurrentVelocity
}

// SetTargetPosition sets the target position (for external control)
func (s *StepMotor) SetTargetPosition(position float64) {
	s.TargetPosition = position
}

// SetEnabled sets the motor enabled state
func (s *StepMotor) SetEnabled(enabled bool) {
	s.Enabled = enabled
}

// Reset resets the motor to initial state
func (s *StepMotor) Reset() {
	s.BaseSensor.Reset()
	s.CurrentPosition = 0
	s.CurrentVelocity = 0
	s.TargetPosition = 0
}
