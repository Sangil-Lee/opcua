package sensors

import (
	"math"
	"time"
)

// ServoMotor simulates a servo motor with velocity/torque control
// Outputs current velocity (RPM) or position (degrees)
type ServoMotor struct {
	BaseSensor
	MaxVelocity     float64 // Maximum velocity (RPM)
	MaxTorque       float64 // Maximum torque (Nm)
	Inertia         float64 // Rotor inertia (kg·m²)
	Damping         float64 // Damping coefficient
	CurrentPosition float64 // Current position (degrees)
	CurrentVelocity float64 // Current velocity (RPM)
	CurrentTorque   float64 // Current torque (Nm)
	TargetVelocity  float64 // Target velocity (RPM)
	LoadTorque      float64 // External load torque (Nm)
	OutputMode      string  // "velocity" or "position"
	AutoMode        bool    // Auto mode for demo
	AutoPattern     string  // Pattern: "sine", "ramp", "step"
	AutoPeriod      float64 // Period for auto pattern (seconds)
	Enabled         bool    // Motor enabled state

	// PID controller for velocity control
	Kp              float64 // Proportional gain
	Ki              float64 // Integral gain
	Kd              float64 // Derivative gain
	integralError   float64 // Accumulated integral error
	lastError       float64 // Last error for derivative
}

// NewServoMotor creates a new servo motor
func NewServoMotor(name, address string, enabled bool, updateIntervalMs int,
	maxVelocity, maxTorque, inertia, damping float64,
	outputMode string, autoMode bool, autoPattern string, autoPeriod float64,
	kp, ki, kd float64, description string) *ServoMotor {
	return &ServoMotor{
		BaseSensor: BaseSensor{
			Name:             name,
			Address:          address,
			Enabled:          enabled,
			UpdateIntervalMs: updateIntervalMs,
			Description:      description,
			ElapsedTime:      0,
		},
		MaxVelocity:     maxVelocity,
		MaxTorque:       maxTorque,
		Inertia:         inertia,
		Damping:         damping,
		CurrentPosition: 0,
		CurrentVelocity: 0,
		CurrentTorque:   0,
		TargetVelocity:  0,
		LoadTorque:      0,
		OutputMode:      outputMode,
		AutoMode:        autoMode,
		AutoPattern:     autoPattern,
		AutoPeriod:      autoPeriod,
		Kp:              kp,
		Ki:              ki,
		Kd:              kd,
		integralError:   0,
		lastError:       0,
	}
}

// Update calculates the current motor state
func (s *ServoMotor) Update(deltaTime time.Duration) float64 {
	if !s.IsEnabled() {
		// Motor disabled, apply only damping
		s.CurrentTorque = 0
		s.ApplyDynamics(deltaTime.Seconds())

		if s.OutputMode == "position" {
			return s.CurrentPosition
		}
		return s.CurrentVelocity
	}

	s.AddElapsedTime(deltaTime)
	elapsed := s.GetElapsedTime()
	dt := deltaTime.Seconds()

	// Auto mode: automatically change target velocity
	if s.AutoMode {
		switch s.AutoPattern {
		case "sine":
			// Sinusoidal velocity variation
			s.TargetVelocity = s.MaxVelocity * 0.7 * math.Sin(2.0*math.Pi*elapsed/s.AutoPeriod)

		case "ramp":
			// Ramp up and down
			progress := math.Mod(elapsed, s.AutoPeriod) / s.AutoPeriod
			if progress < 0.5 {
				s.TargetVelocity = s.MaxVelocity * (progress * 2.0)
			} else {
				s.TargetVelocity = s.MaxVelocity * (2.0 - progress*2.0)
			}

		case "step":
			// Step pattern
			stepIndex := int(elapsed/s.AutoPeriod) % 3
			velocities := []float64{0, s.MaxVelocity * 0.5, s.MaxVelocity * 0.8}
			s.TargetVelocity = velocities[stepIndex]
		}
	}

	// PID velocity control
	velocityError := s.TargetVelocity - s.CurrentVelocity

	// Proportional term
	pTerm := s.Kp * velocityError

	// Integral term (with anti-windup)
	s.integralError += velocityError * dt
	maxIntegral := s.MaxTorque / s.Ki
	if s.integralError > maxIntegral {
		s.integralError = maxIntegral
	} else if s.integralError < -maxIntegral {
		s.integralError = -maxIntegral
	}
	iTerm := s.Ki * s.integralError

	// Derivative term
	dTerm := 0.0
	if dt > 0 {
		dTerm = s.Kd * (velocityError - s.lastError) / dt
	}
	s.lastError = velocityError

	// Calculate motor torque
	s.CurrentTorque = pTerm + iTerm + dTerm

	// Clamp torque to maximum
	if s.CurrentTorque > s.MaxTorque {
		s.CurrentTorque = s.MaxTorque
	} else if s.CurrentTorque < -s.MaxTorque {
		s.CurrentTorque = -s.MaxTorque
	}

	// Apply dynamics
	s.ApplyDynamics(dt)

	// Return output based on mode
	if s.OutputMode == "position" {
		return s.CurrentPosition
	}
	return s.CurrentVelocity
}

// ApplyDynamics applies motor dynamics (torque -> acceleration -> velocity -> position)
func (s *ServoMotor) ApplyDynamics(dt float64) {
	// Net torque = motor torque - load torque - damping torque
	netTorque := s.CurrentTorque - s.LoadTorque - s.Damping*s.CurrentVelocity

	// Angular acceleration (rad/s²)
	// Convert RPM to rad/s: ω = RPM × 2π/60
	omegaRadPerSec := s.CurrentVelocity * 2.0 * math.Pi / 60.0

	// τ = I × α  =>  α = τ / I
	if s.Inertia > 0 {
		angularAccel := netTorque / s.Inertia // rad/s²
		omegaRadPerSec += angularAccel * dt

		// Convert back to RPM
		s.CurrentVelocity = omegaRadPerSec * 60.0 / (2.0 * math.Pi)
	}

	// Clamp velocity to maximum
	if s.CurrentVelocity > s.MaxVelocity {
		s.CurrentVelocity = s.MaxVelocity
	} else if s.CurrentVelocity < -s.MaxVelocity {
		s.CurrentVelocity = -s.MaxVelocity
	}

	// Update position (degrees)
	// θ = ∫ω dt, where ω is in RPM
	// Convert RPM to degrees/second: deg/s = RPM × 360 / 60 = RPM × 6
	s.CurrentPosition += s.CurrentVelocity * 6.0 * dt

	// Wrap position to [0, 360)
	s.CurrentPosition = math.Mod(s.CurrentPosition, 360.0)
	if s.CurrentPosition < 0 {
		s.CurrentPosition += 360.0
	}
}

// GetPosition returns the current position (for monitoring)
func (s *ServoMotor) GetPosition() float64 {
	return s.CurrentPosition
}

// GetVelocity returns the current velocity (for monitoring)
func (s *ServoMotor) GetVelocity() float64 {
	return s.CurrentVelocity
}

// GetTorque returns the current torque (for monitoring)
func (s *ServoMotor) GetTorque() float64 {
	return s.CurrentTorque
}

// SetTargetVelocity sets the target velocity (for external control)
func (s *ServoMotor) SetTargetVelocity(velocity float64) {
	s.TargetVelocity = velocity
}

// SetLoadTorque sets the external load torque
func (s *ServoMotor) SetLoadTorque(torque float64) {
	s.LoadTorque = torque
}

// SetEnabled sets the motor enabled state
func (s *ServoMotor) SetEnabled(enabled bool) {
	s.Enabled = enabled
}

// Reset resets the motor to initial state
func (s *ServoMotor) Reset() {
	s.BaseSensor.Reset()
	s.CurrentPosition = 0
	s.CurrentVelocity = 0
	s.CurrentTorque = 0
	s.TargetVelocity = 0
	s.integralError = 0
	s.lastError = 0
}
