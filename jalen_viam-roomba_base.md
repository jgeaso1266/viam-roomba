# Model jalen:viam-roomba:base

A Viam base component for the iRobot Roomba 650/655 using the Roomba Open Interface (OI) serial protocol. Supports full movement control via `SetVelocity`, `SetPower`, `MoveStraight`, and `Spin`, with accurate movement detection by polling wheel velocity sensors.

## Configuration

```json
{
  "serial_port": "<string>",
  "width_mm": <int>,
  "wheel_circumference_mm": <int>
}
```

### Attributes

| Name                    | Type   | Inclusion | Description                                                                 |
|-------------------------|--------|-----------|-----------------------------------------------------------------------------|
| `serial_port`           | string | Required  | Serial port path for the USB-to-TTL adapter (e.g. `/dev/ttyUSB0`)          |
| `width_mm`              | int    | Optional  | Wheelbase width in mm. Defaults to `235` (Roomba 600 series)                |
| `wheel_circumference_mm`| int    | Optional  | Wheel circumference in mm. Defaults to `220` (Roomba 600 series)            |

### Example Configuration

```json
{
  "serial_port": "/dev/ttyUSB0",
  "width_mm": 235,
  "wheel_circumference_mm": 220
}
```

## DoCommand

### `enter_full_mode`

Switches the Roomba to Full mode, disabling cliff sensor and other safety stops. Useful when the robot is on an elevated surface or in a test environment.

```json
{ "command": "enter_full_mode" }
```

### `enter_safe_mode`

Switches the Roomba back to Safe mode, re-enabling safety features.

```json
{ "command": "enter_safe_mode" }
```

### `seek_dock`

Sends the Roomba to its charging dock.

```json
{ "command": "seek_dock" }
```

### `clean`

Starts the Roomba's default cleaning routine.

```json
{ "command": "clean" }
```

### `stop`

Immediately stops all wheel movement.

```json
{ "command": "stop" }
```
