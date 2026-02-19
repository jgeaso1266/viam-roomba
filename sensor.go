package viamroomba

import (
	"context"
	"encoding/binary"
	"fmt"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var Sensor = resource.NewModel("jalen", "viam-roomba", "sensor")

func init() {
	resource.RegisterComponent(sensor.API, Sensor,
		resource.Registration[sensor.Sensor, *SensorConfig]{
			Constructor: newViamRoombaSensor,
		},
	)
}

type SensorConfig struct {
	SerialPort string `json:"serial_port"`
}

func (cfg *SensorConfig) Validate(path string) ([]string, []string, error) {
	if cfg.SerialPort == "" {
		return nil, nil, fmt.Errorf("%s: serial_port is required", path)
	}
	return nil, nil, nil
}

type viamRoombaSensor struct {
	resource.AlwaysRebuild

	name       resource.Name
	logger     logging.Logger
	conn       *roombaConn
	serialPort string
}

func newViamRoombaSensor(ctx context.Context, deps resource.Dependencies, rawConf resource.Config, logger logging.Logger) (sensor.Sensor, error) {
	conf, err := resource.NativeConfig[*SensorConfig](rawConf)
	if err != nil {
		return nil, err
	}

	conn, err := acquireConn(conf.SerialPort)
	if err != nil {
		return nil, err
	}

	logger.Infof("Roomba sensor initialized on %s", conf.SerialPort)

	return &viamRoombaSensor{
		name:       rawConf.ResourceName(),
		logger:     logger,
		conn:       conn,
		serialPort: conf.SerialPort,
	}, nil
}

func (s *viamRoombaSensor) Name() resource.Name {
	return s.name
}

// sensorPackets lists all queried packet IDs in order. Index in this slice
// corresponds to the index in the data slice returned by QueryList.
var sensorPackets = []byte{
	7,  // Bumps and Wheel Drops
	8,  // Wall
	9,  // Cliff Left
	10, // Cliff Front Left
	11, // Cliff Front Right
	12, // Cliff Right
	13, // Virtual Wall
	14, // Overcurrents
	15, // Dirt Detect
	17, // IR Opcode
	18, // Buttons
	19, // Distance (mm, signed)
	20, // Angle (degrees, signed)
	21, // Charging State
	22, // Voltage (mV)
	23, // Current (mA, signed)
	24, // Temperature (°C, signed)
	25, // Battery Charge (mAh)
	26, // Battery Capacity (mAh)
	27, // Wall Signal
	28, // Cliff Left Signal
	29, // Cliff Front Left Signal
	30, // Cliff Front Right Signal
	31, // Cliff Right Signal
	34, // Charging Sources Available
	35, // OI Mode
	39, // Requested Velocity (mm/s, signed)
	40, // Requested Radius (mm, signed)
}

var chargingStates = []string{"not_charging", "reconditioning", "full_charging", "trickle_charging", "waiting", "charging_fault"}
var oiModes = []string{"off", "passive", "safe", "full"}

func (s *viamRoombaSensor) Readings(ctx context.Context, extra map[string]any) (map[string]any, error) {
	s.conn.mu.Lock()
	defer s.conn.mu.Unlock()

	s.conn.flushRx()
	data, err := s.conn.roomba.QueryList(sensorPackets)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensors: %w", err)
	}
	if len(data) != len(sensorPackets) {
		return nil, fmt.Errorf("unexpected sensor data count: got %d, want %d", len(data), len(sensorPackets))
	}

	b := func(idx int) byte { return data[idx][0] }
	i16 := func(idx int) int16 { return int16(binary.BigEndian.Uint16(data[idx])) }
	u16 := func(idx int) uint16 { return binary.BigEndian.Uint16(data[idx]) }

	readings := map[string]any{}

	// Packet 7: Bumps and Wheel Drops
	bumps := b(0)
	readings["bump_right"] = bumps&0x01 != 0
	readings["bump_left"] = bumps&0x02 != 0
	readings["wheel_drop_right"] = bumps&0x04 != 0
	readings["wheel_drop_left"] = bumps&0x08 != 0

	// Packets 8-12: Proximity sensors
	readings["wall"] = b(1)&0x01 != 0
	readings["cliff_left"] = b(2)&0x01 != 0
	readings["cliff_front_left"] = b(3)&0x01 != 0
	readings["cliff_front_right"] = b(4)&0x01 != 0
	readings["cliff_right"] = b(5)&0x01 != 0

	// Packet 13: Virtual Wall
	readings["virtual_wall"] = b(6)&0x01 != 0

	// Packet 14: Overcurrents (bitmask: side brush, main brush, right wheel, left wheel)
	overcurrents := b(7)
	readings["overcurrent_side_brush"] = overcurrents&0x01 != 0
	readings["overcurrent_main_brush"] = overcurrents&0x04 != 0
	readings["overcurrent_right_wheel"] = overcurrents&0x08 != 0
	readings["overcurrent_left_wheel"] = overcurrents&0x10 != 0

	// Packet 15: Dirt Detect
	readings["dirt_detect"] = int(b(8))

	// Packet 17: IR Opcode
	readings["ir_opcode"] = int(b(9))

	// Packet 18: Buttons
	buttons := b(10)
	readings["button_clean"] = buttons&0x01 != 0
	readings["button_spot"] = buttons&0x02 != 0
	readings["button_dock"] = buttons&0x04 != 0
	readings["button_minute"] = buttons&0x08 != 0
	readings["button_hour"] = buttons&0x10 != 0
	readings["button_day"] = buttons&0x20 != 0
	readings["button_schedule"] = buttons&0x40 != 0
	readings["button_clock"] = buttons&0x80 != 0

	// Packets 19-20: Odometry (cumulative since last read)
	readings["distance_mm"] = int(i16(11))
	readings["angle_deg"] = int(i16(12))

	// Packet 21: Charging State
	chargingIdx := int(b(13))
	if chargingIdx < len(chargingStates) {
		readings["charging_state"] = chargingStates[chargingIdx]
	} else {
		readings["charging_state"] = "unknown"
	}

	// Packets 22-26: Battery
	readings["voltage_mv"] = int(u16(14))
	readings["current_ma"] = int(i16(15))
	readings["temperature_c"] = int(int8(b(16)))
	charge := int(u16(17))
	capacity := int(u16(18))
	readings["battery_charge_mah"] = charge
	readings["battery_capacity_mah"] = capacity
	if capacity > 0 {
		readings["battery_percent"] = float64(charge) / float64(capacity) * 100.0
	}

	// Packets 27-31: Signal strengths
	readings["wall_signal"] = int(u16(19))
	readings["cliff_left_signal"] = int(u16(20))
	readings["cliff_front_left_signal"] = int(u16(21))
	readings["cliff_front_right_signal"] = int(u16(22))
	readings["cliff_right_signal"] = int(u16(23))

	// Packet 33: Charging Sources Available
	charger := b(24)
	readings["charger_internal"] = charger&0x01 != 0
	readings["charger_homebase"] = charger&0x02 != 0

	// Packet 34: OI Mode
	modeIdx := int(b(25))
	if modeIdx < len(oiModes) {
		readings["oi_mode"] = oiModes[modeIdx]
	} else {
		readings["oi_mode"] = "unknown"
	}

	// Packets 39-40: Requested motion
	readings["requested_velocity_mms"] = int(i16(26))
	readings["requested_radius_mm"] = int(i16(27))

	return readings, nil
}

func (s *viamRoombaSensor) DoCommand(ctx context.Context, cmd map[string]any) (map[string]any, error) {
	return nil, nil
}

func (s *viamRoombaSensor) Close(ctx context.Context) error {
	releaseConn(s.serialPort)
	return nil
}
