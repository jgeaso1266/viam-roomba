package viamroomba

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/golang/geo/r3"
	"github.com/parabolala/go-roomba"
	base "go.viam.com/rdk/components/base"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/operation"
	"go.viam.com/rdk/resource"
	"go.viam.com/rdk/spatialmath"
)

var (
	Base             = resource.NewModel("jalen", "viam-roomba", "base")
	errUnimplemented = errors.New("unimplemented")
)

func init() {
	resource.RegisterComponent(base.API, Base,
		resource.Registration[base.Base, *Config]{
			Constructor: newViamRoombaBase,
		},
	)
}

type Config struct {
	SerialPort           string `json:"serial_port"`
	WidthMM              int    `json:"width_mm,omitempty"`
	WheelCircumferenceMM int    `json:"wheel_circumference_mm,omitempty"`
}

func (cfg *Config) Validate(path string) ([]string, []string, error) {
	if cfg.SerialPort == "" {
		return nil, nil, fmt.Errorf("%s: serial_port is required", path)
	}

	if cfg.WidthMM < 0 {
		return nil, nil, fmt.Errorf("%s: width_mm must be a positive number", path)
	}
	if cfg.WheelCircumferenceMM < 0 {
		return nil, nil, fmt.Errorf("%s: wheel_circumference_mm must be a positive number", path)
	}

	return nil, nil, nil
}

type viamRoombaBase struct {
	resource.AlwaysRebuild

	name   resource.Name
	logger logging.Logger
	cfg    *Config

	// Roomba hardware connection
	roomba *roomba.Roomba

	// Robot properties
	widthMM              int
	wheelCircumferenceMM int

	// Operation management
	opMgr *operation.SingleOperationManager

	cancelCtx  context.Context
	cancelFunc func()
	mu         sync.Mutex
}

func newViamRoombaBase(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (base.Base, error) {
	conf, err := resource.NativeConfig[*Config](rawConf)
	if err != nil {
		return nil, err
	}

	return NewBase(ctx, deps, rawConf.ResourceName(), conf, logger)

}

func NewBase(ctx context.Context, deps resource.Dependencies, name resource.Name, conf *Config, logger logging.Logger) (base.Base, error) {
	cancelCtx, cancelFunc := context.WithCancel(context.Background())

	// Initialize Roomba connection
	r, err := roomba.MakeRoomba(conf.SerialPort)
	if err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to connect to Roomba on %s: %w", conf.SerialPort, err)
	}

	// Enter Safe mode (enables control with safety features)
	if err := r.Safe(); err != nil {
		cancelFunc()
		return nil, fmt.Errorf("failed to enter Safe mode: %w", err)
	}

	widthMM := conf.WidthMM
	if widthMM == 0 {
		widthMM = 235
	}
	wheelCircumferenceMM := conf.WheelCircumferenceMM
	if wheelCircumferenceMM == 0 {
		wheelCircumferenceMM = 220
	}

	s := &viamRoombaBase{
		name:                 name,
		logger:               logger,
		cfg:                  conf,
		roomba:               r,
		widthMM:              widthMM,
		wheelCircumferenceMM: wheelCircumferenceMM,
		opMgr:                operation.NewSingleOperationManager(),
		cancelCtx:            cancelCtx,
		cancelFunc:           cancelFunc,
	}

	logger.Infof("Roomba base initialized on %s (width: %dmm, wheel circumference: %dmm)",
		conf.SerialPort, widthMM, wheelCircumferenceMM)

	return s, nil
}

func (s *viamRoombaBase) Name() resource.Name {
	return s.name
}

// MoveStraight moves the robot straight a given distance at a given speed.
// If a distance or speed of zero is given, the base will stop.
// This method blocks until completed or cancelled.
func (s *viamRoombaBase) MoveStraight(ctx context.Context, distanceMm int, mmPerSec float64, extra map[string]interface{}) error {
	ctx, done := s.opMgr.New(ctx)
	defer done()

	// If distance or speed is zero, stop
	if distanceMm == 0 || mmPerSec == 0 {
		return s.Stop(ctx, extra)
	}

	// Calculate expected duration
	duration := math.Abs(float64(distanceMm) / mmPerSec)

	// Determine direction and velocity
	var velocity int16
	if distanceMm > 0 {
		velocity = int16(mmPerSec)
	} else {
		velocity = -int16(mmPerSec)
	}

	// Clamp velocity to Roomba limits
	if velocity > 500 {
		velocity = 500
	} else if velocity < -500 {
		velocity = -500
	}

	s.mu.Lock()
	if err := s.roomba.Drive(velocity, 32767); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to start straight movement: %w", err)
	}
	s.mu.Unlock()

	s.logger.Debugf("MoveStraight: distance=%d mm, velocity=%d mm/sec, duration=%.2f sec", distanceMm, velocity, duration)

	// Block for calculated duration or until context is cancelled
	select {
	case <-ctx.Done():
		s.Stop(ctx, extra)
		return ctx.Err()
	case <-s.cancelCtx.Done():
		s.Stop(ctx, extra)
		return s.cancelCtx.Err()
	case <-context.WithValue(ctx, "timer", true).Done():
		// This will never fire, we use sleep below instead
	}

	// Sleep for the calculated duration
	sleepCtx, cancel := context.WithTimeout(ctx, time.Duration(duration*1000)*time.Millisecond)
	defer cancel()

	<-sleepCtx.Done()

	// Stop when complete
	return s.Stop(ctx, extra)
}

// Spin spins the robot by a given angle in degrees at a given speed.
// If a speed of 0 the base will stop.
// Given a positive speed and a positive angle, the base turns to the left (for built-in RDK drivers).
// This method blocks until completed or cancelled.
func (s *viamRoombaBase) Spin(ctx context.Context, angleDeg float64, degsPerSec float64, extra map[string]interface{}) error {
	ctx, done := s.opMgr.New(ctx)
	defer done()

	// If angle or speed is zero, stop
	if angleDeg == 0 || degsPerSec == 0 {
		return s.Stop(ctx, extra)
	}

	// Calculate expected duration
	duration := math.Abs(angleDeg / degsPerSec)

	// Determine direction for spin in place
	var radius int16
	if angleDeg > 0 {
		// Left turn (positive angle)
		radius = 1 // Spin in place CCW
	} else {
		// Right turn (negative angle)
		radius = -1 // Spin in place CW
	}

	s.mu.Lock()
	// For spin in place, velocity can be set to a small value
	// The radius (1 or -1) determines direction
	if err := s.roomba.Drive(100, radius); err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to start spin: %w", err)
	}
	s.mu.Unlock()

	s.logger.Debugf("Spin: angle=%.2f deg, speed=%.2f deg/sec, duration=%.2f sec", angleDeg, degsPerSec, duration)

	// Block for calculated duration or until context is cancelled
	timer := context.WithValue(ctx, "duration", duration)
	sleepCtx, cancel := context.WithTimeout(timer, time.Duration(duration*1000)*time.Millisecond)
	defer cancel()

	select {
	case <-sleepCtx.Done():
		// Timer expired, movement complete
	case <-ctx.Done():
		s.Stop(ctx, extra)
		return ctx.Err()
	case <-s.cancelCtx.Done():
		s.Stop(ctx, extra)
		return s.cancelCtx.Err()
	}

	// Stop when complete
	return s.Stop(ctx, extra)
}

// Set the power of the base.
// For linear power, positive Y moves forwards for built-in RDK drivers.
// For angular power, positive Z turns to the left for built-in RDK drivers.
func (s *viamRoombaBase) SetPower(ctx context.Context, linear r3.Vector, angular r3.Vector, extra map[string]interface{}) error {
	const maxWheelSpeed = 500.0
	// Derive max angular velocity so full power produces the same wheel speed as full linear power
	maxAngularDegPerSec := maxWheelSpeed * 180.0 / (math.Pi * float64(s.widthMM) / 2.0)

	linearVel := r3.Vector{X: 0, Y: linear.Y * maxWheelSpeed, Z: 0}
	angularVel := r3.Vector{X: 0, Y: 0, Z: angular.Z * maxAngularDegPerSec}

	return s.SetVelocity(ctx, linearVel, angularVel, extra)
}

// Set the velocity of the base.
// linear is in mmPerSec (positive Y moves forwards for built-in RDK drivers).
// angular is in degsPerSec (positive Z turns to the left for built-in RDK drivers).
func (s *viamRoombaBase) SetVelocity(ctx context.Context, linear r3.Vector, angular r3.Vector, extra map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If both linear and angular are zero, stop
	if linear.Y == 0 && angular.Z == 0 {
		return s.roomba.Stop()
	}

	linearMM := linear.Y
	angularVel := angular.Z

	var velocity int16
	var radius int16

	if linearMM == 0 && angularVel != 0 {
		angularRadPerSec := math.Abs(angularVel) * math.Pi / 180.0
		wheelSpeed := angularRadPerSec * float64(s.widthMM) / 2.0
		velocity = int16(math.Min(500, wheelSpeed))
		if angularVel > 0 {
			radius = 1
		} else {
			radius = -1
		}
	} else {
		velocity = int16(linearMM)
		if velocity > 500 {
			s.logger.Warnf("Clamping velocity from %d to 500 mm/sec", velocity)
			velocity = 500
		} else if velocity < -500 {
			s.logger.Warnf("Clamping velocity from %d to -500 mm/sec", velocity)
			velocity = -500
		}

		if angularVel == 0 {
			radius = 32767 // Drive straight
		} else {
			// radius = (velocity * 180) / (angularVel * π)
			radiusFloat := (float64(velocity) * 180.0) / (angularVel * math.Pi)
			radius = int16(math.Max(-2000, math.Min(2000, radiusFloat)))
		}
	}

	if err := s.roomba.Drive(velocity, radius); err != nil {
		return fmt.Errorf("failed to drive Roomba: %w", err)
	}

	s.logger.Debugf("SetVelocity: velocity=%d mm/sec, radius=%d mm", velocity, radius)
	return nil
}

func (s *viamRoombaBase) Stop(ctx context.Context, extra map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.roomba.Stop(); err != nil {
		return fmt.Errorf("failed to stop Roomba: %w", err)
	}

	s.logger.Debug("Roomba stopped")
	return nil
}

func (s *viamRoombaBase) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	cmdName, ok := cmd["command"].(string)
	if !ok {
		return nil, fmt.Errorf("command must be a string")
	}

	switch cmdName {
	case "enter_full_mode":
		if err := s.roomba.Full(); err != nil {
			return nil, fmt.Errorf("failed to enter Full mode: %w", err)
		}
		s.logger.Info("Entered Full mode (safety features disabled)")
		return map[string]interface{}{"status": "full_mode_enabled"}, nil

	case "enter_safe_mode":
		if err := s.roomba.Safe(); err != nil {
			return nil, fmt.Errorf("failed to enter Safe mode: %w", err)
		}
		s.logger.Info("Entered Safe mode (safety features enabled)")
		return map[string]interface{}{"status": "safe_mode_enabled"}, nil

	case "seek_dock":
		if err := s.roomba.SeekDock(); err != nil {
			return nil, fmt.Errorf("failed to seek dock: %w", err)
		}
		s.logger.Info("Seeking charging dock")
		return map[string]interface{}{"status": "seeking_dock"}, nil

	case "clean":
		if err := s.roomba.Clean(); err != nil {
			return nil, fmt.Errorf("failed to start cleaning: %w", err)
		}
		s.logger.Info("Started cleaning mode")
		return map[string]any{"status": "cleaning"}, nil

	case "stop":
		if err := s.roomba.Stop(); err != nil {
			return nil, fmt.Errorf("failed to stop: %w", err)
		}
		return map[string]any{"status": "stopped"}, nil

	default:
		return nil, fmt.Errorf("unknown command: %s", cmdName)
	}
}

func (s *viamRoombaBase) IsMoving(ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Query Roomba velocity sensors to get actual movement state
	// Sensor packet 39: Left wheel velocity (int16, mm/s)
	// Sensor packet 40: Right wheel velocity (int16, mm/s)

	// Request left wheel velocity
	leftVel, err := s.roomba.Sensors(39)
	if err != nil {
		return false, fmt.Errorf("failed to read left wheel velocity: %w", err)
	}

	// Request right wheel velocity
	rightVel, err := s.roomba.Sensors(40)
	if err != nil {
		return false, fmt.Errorf("failed to read right wheel velocity: %w", err)
	}

	// Parse velocities (2 bytes each, signed int16, big-endian)
	if len(leftVel) < 2 || len(rightVel) < 2 {
		return false, fmt.Errorf("invalid sensor data length")
	}

	leftSpeed := int16(binary.BigEndian.Uint16(leftVel))
	rightSpeed := int16(binary.BigEndian.Uint16(rightVel))

	// Consider moving if either wheel has non-zero velocity
	// Use small threshold to handle sensor noise
	threshold := int16(5) // 5 mm/s threshold
	isMoving := math.Abs(float64(leftSpeed)) > float64(threshold) ||
		math.Abs(float64(rightSpeed)) > float64(threshold)

	s.logger.Debugf("IsMoving: left=%d mm/s, right=%d mm/s, moving=%v", leftSpeed, rightSpeed, isMoving)
	return isMoving, nil
}

// Properties returns the width, turning radius, and wheel circumference of the physical base in meters.
func (s *viamRoombaBase) Properties(ctx context.Context, extra map[string]interface{}) (base.Properties, error) {
	return base.Properties{
		WidthMeters:              float64(s.widthMM) / 1000.0,
		TurningRadiusMeters:      0.0, // Differential drive can turn in place
		WheelCircumferenceMeters: float64(s.wheelCircumferenceMM) / 1000.0,
	}, nil
}

func (s *viamRoombaBase) Geometries(ctx context.Context, extra map[string]any) ([]spatialmath.Geometry, error) {
	// Roomba 650: 340mm diameter, 92mm height. Sphere approximation preserves the circular footprint.
	geom, err := spatialmath.NewSphere(spatialmath.NewZeroPose(), 170.0, s.name.Name)
	if err != nil {
		return nil, err
	}
	return []spatialmath.Geometry{geom}, nil
}

func (s *viamRoombaBase) Close(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Stop the Roomba
	if err := s.roomba.Stop(); err != nil {
		s.logger.Warnf("Failed to stop Roomba during close: %v", err)
	}

	// Cancel context
	s.cancelFunc()

	s.logger.Info("Roomba base closed")
	return nil
}
