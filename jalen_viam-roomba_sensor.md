# Model jalen:viam-roomba:sensor

A Viam sensor component for the iRobot Roomba 650/655 that exposes Roomba Open Interface sensor data. Returns real-time readings for bump sensors, cliff sensors, battery state, OI mode, and more.

## Configuration

```json
{
  "serial_port": "<string>"
}
```

### Attributes

| Name          | Type   | Inclusion | Description                                                        |
|---------------|--------|-----------|--------------------------------------------------------------------|
| `serial_port` | string | Required  | Serial port path for the USB-to-TTL adapter (e.g. `/dev/ttyUSB0`) |

### Example Configuration

```json
{
  "serial_port": "/dev/ttyUSB0"
}
```

> **Note:** When running alongside the `jalen:viam-roomba:base` component on the same serial port, the two components share the underlying connection. The base component owns mode initialization (Safe/Full mode); the sensor component reads data without changing the OI mode.

## Readings

| Key                        | Type    | Description                                          |
|----------------------------|---------|------------------------------------------------------|
| `bump_right`               | bool    | Right bumper pressed                                 |
| `bump_left`                | bool    | Left bumper pressed                                  |
| `wheel_drop_right`         | bool    | Right wheel dropped                                  |
| `wheel_drop_left`          | bool    | Left wheel dropped                                   |
| `wall`                     | bool    | Wall sensor active                                   |
| `cliff_left`               | bool    | Cliff detected on left                               |
| `cliff_front_left`         | bool    | Cliff detected on front-left                         |
| `cliff_front_right`        | bool    | Cliff detected on front-right                        |
| `cliff_right`              | bool    | Cliff detected on right                              |
| `virtual_wall`             | bool    | Virtual wall signal detected                         |
| `overcurrent_side_brush`   | bool    | Side brush overcurrent                               |
| `overcurrent_main_brush`   | bool    | Main brush overcurrent                               |
| `overcurrent_right_wheel`  | bool    | Right wheel overcurrent                              |
| `overcurrent_left_wheel`   | bool    | Left wheel overcurrent                               |
| `dirt_detect`              | int     | Dirt detect sensor level (0–255)                     |
| `ir_opcode`                | int     | IR opcode received from remote or dock               |
| `button_clean`             | bool    | Clean button pressed                                 |
| `button_spot`              | bool    | Spot button pressed                                  |
| `button_dock`              | bool    | Dock button pressed                                  |
| `button_minute`            | bool    | Minute button pressed                                |
| `button_hour`              | bool    | Hour button pressed                                  |
| `button_day`               | bool    | Day button pressed                                   |
| `button_schedule`          | bool    | Schedule button pressed                              |
| `button_clock`             | bool    | Clock button pressed                                 |
| `distance_mm`              | int     | Distance traveled since last read (mm, signed)       |
| `angle_deg`                | int     | Angle turned since last read (degrees, signed)       |
| `charging_state`           | string  | One of: `not_charging`, `reconditioning`, `full_charging`, `trickle_charging`, `waiting`, `charging_fault` |
| `voltage_mv`               | int     | Battery voltage (mV)                                 |
| `current_ma`               | int     | Battery current draw (mA, negative = discharging)    |
| `temperature_c`            | int     | Battery temperature (°C)                             |
| `battery_charge_mah`       | int     | Battery charge remaining (mAh)                       |
| `battery_capacity_mah`     | int     | Battery total capacity (mAh)                         |
| `battery_percent`          | float   | Battery charge percentage (only present if capacity > 0) |
| `wall_signal`              | int     | Wall sensor signal strength (0–4095)                 |
| `cliff_left_signal`        | int     | Cliff left sensor signal strength (0–4095)           |
| `cliff_front_left_signal`  | int     | Cliff front-left sensor signal strength (0–4095)     |
| `cliff_front_right_signal` | int     | Cliff front-right sensor signal strength (0–4095)    |
| `cliff_right_signal`       | int     | Cliff right sensor signal strength (0–4095)          |
| `charger_internal`         | bool    | Internal charger present                             |
| `charger_homebase`         | bool    | Home base charger present                            |
| `oi_mode`                  | string  | Current OI mode: `off`, `passive`, `safe`, or `full` |
| `requested_velocity_mms`   | int     | Last commanded velocity (mm/s, signed)               |
| `requested_radius_mm`      | int     | Last commanded radius (mm, signed)                   |
