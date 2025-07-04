package message

// DataType represents the type of sensor data being transmitted.
type DataType int8

// Status represents the operational status of a sensor device.
type Status uint8

// ConfigKey represents configuration parameter identifiers.
type ConfigKey uint8

// CustomType represents the type of custom data payload.
type CustomType uint8

// AckStatus represents acknowledgment response status codes.
type AckStatus uint8

// DeviceType represents the hardware type of the sensor device.
type DeviceType uint8

// Sensor data type constants defining supported sensor measurements.
const (
	Accelerometer DataType = 0x01 // 3-axis accelerometer data (m/s²)
	Gyroscope     DataType = 0x02 // 3-axis gyroscope data (rad/s)
	Quaternion    DataType = 0x03 // 4-component quaternion orientation
	EulerAngles   DataType = 0x04 // 3-axis Euler angles (roll, pitch, yaw)
)

// Device status constants indicating operational state.
const (
	Ok          Status = 0x01 // Device operating normally
	Expectation Status = 0x02 // Device waiting for input/command
	Collection  Status = 0x03 // Device actively collecting data
	LowBattery  Status = 0x04 // Device battery level is low
	Error       Status = 0x05 // Device encountered an error
)

// Configuration parameter key constants for device settings.
const (
	ConfigKeySampleRate  ConfigKey = 0x01 // Data sampling rate (Hz)
	ConfigKeyRange       ConfigKey = 0x02 // Sensor measurement range
	ConfigKeyMAC         ConfigKey = 0x03 // Device MAC address
	ConfigKeyDeviceName  ConfigKey = 0x04 // Human-readable device name
	ConfigKeyIPAddress   ConfigKey = 0x05 // Network IP address
	ConfigKeyMode        ConfigKey = 0x06 // Operating mode setting
	ConfigKeySensitivity ConfigKey = 0x07 // Sensor sensitivity level
	ConfigKeyCalibration ConfigKey = 0x08 // Calibration parameters
)

// Custom data type constants for application-specific payloads.
const (
	CustomTypeLog    CustomType = 0x01 // Log message data
	CustomTypeError  CustomType = 0x02 // Error report data
	CustomTypeDebug  CustomType = 0x03 // Debug information
	CustomTypeString CustomType = 0x04 // String/text data
	CustomTypeBinary CustomType = 0x05 // Binary blob data
)

// Acknowledgment status constants for message responses.
const (
	AckOK             AckStatus = 0x01 // Message processed successfully
	AckError          AckStatus = 0x02 // General processing error
	AckInvalidCRC     AckStatus = 0x03 // CRC validation failed
	AckUnknownMessage AckStatus = 0x04 // Unrecognized message type
	AckBufferFull     AckStatus = 0x05 // Receive buffer overflow
)

// Device type constants identifying hardware capabilities.
const (
	DeviceType3Axis  DeviceType = 0x01 // 3-axis sensor (accelerometer only)
	DeviceType6Axis  DeviceType = 0x02 // 6-axis sensor (accel + gyro)
	DeviceType9Axis  DeviceType = 0x03 // 9-axis sensor (accel + gyro + mag)
	DeviceTypeHub    DeviceType = 0x10 // Central hub device
	DeviceTypeRelay  DeviceType = 0x11 // Message relay device
	DeviceTypeCustom DeviceType = 0xFF // Custom/application-specific device
)

// Device capability flags for sensor features (bitmask).
const (
	CapAccelerometer uint8 = 1 << 0 // 3-axis acceleration measurement
	CapGyroscope     uint8 = 1 << 1 // 3-axis angular velocity measurement
	CapMagnetometer  uint8 = 1 << 2 // 3-axis magnetic field measurement
	CapQuaternion    uint8 = 1 << 3 // Quaternion orientation calculation
	CapTemperature   uint8 = 1 << 4 // Temperature measurement
)

// Item represents a configuration key-value pair with length prefix.
type Item struct {
	Key    ConfigKey // Configuration parameter identifier
	Length uint8     // Length of the value in bytes
	Value  []byte    // Configuration value data
}

// SensorCommand represents a command sent to a specific sensor device.
type SensorCommand struct {
	SensorID  uint8  // Target sensor identifier
	TimeStamp uint32 // Command timestamp (Unix seconds)
	Command   uint8  // Command code to execute
}

// SensorConfig represents configuration parameters for a sensor device.
type SensorConfig struct {
	SensorID  uint8  // Target sensor identifier
	TimeStamp uint32 // Configuration timestamp (Unix seconds)
	Config    []Item // Array of configuration key-value pairs
}

// SensorHeartbeat represents periodic status information from a sensor.
type SensorHeartbeat struct {
	SensorID  uint8  // Source sensor identifier
	TimeStamp uint32 // Heartbeat timestamp (Unix seconds)
	Battery   uint8  // Battery level percentage (0-100)
	Status    Status // Current operational status
}

// SensorData represents a single sensor measurement.
type SensorData struct {
	SensorID  uint8  // Source sensor identifier
	TimeStamp uint32 // Measurement timestamp (Unix seconds)
	Data      Data   // Sensor measurement data
}

// CustomData represents application-specific data payload.
type CustomData struct {
	SensorID  uint8      // Source sensor identifier
	TimeStamp uint32     // Data timestamp (Unix seconds)
	DataType  CustomType // Type of custom data
	Data      []Item     // Custom data items
}

// TimeSync represents time synchronization between server and sensor.
type TimeSync struct {
	SensorID   uint8  // Target sensor identifier
	ServerTime uint32 // Server timestamp (Unix seconds)
	SensorTime uint32 // Sensor's current timestamp (Unix seconds)
}

// Ack represents an acknowledgment response to a received message.
type Ack struct {
	SensorID  uint8     // Source sensor identifier
	MessageID uint16    // ID of the message being acknowledged
	Status    AckStatus // Processing status of the acknowledged message
}

// Registration represents initial sensor registration with the system.
type Registration struct {
	SensorID     uint8      // Unique sensor identifier
	DeviceType   DeviceType // Hardware type of the device
	Capabilities uint8      // Bitmask of device capabilities
	FWVersion    uint16     // Firmware version (major.minor format)
}

// Fragment represents a piece of a large message split across multiple packets.
type Fragment struct {
	MessageID      uint16 // Unique identifier for the fragmented message
	FragmentNum    uint8  // Current fragment number (0-based)
	TotalFragments uint8  // Total number of fragments for this message
	Data           []byte // Fragment payload data
}

// RelayedMessage represents a message forwarded through a hub device.
type RelayedMessage struct {
	RelayID      uint8  // Identifier of the relay/hub device
	OriginalData []byte // Original message data being relayed
}

// Data represents a single sensor measurement with type and values.
type Data struct {
	Type   DataType  // Type of sensor data (accelerometer, gyroscope, etc.)
	Values []float32 // Measurement values (typically 3 or 4 components)
}

// SensorDataMulti represents multiple sensor measurements in a single message.
type SensorDataMulti struct {
	SensorID  uint8  // Source sensor identifier
	TimeStamp uint32 // Measurement timestamp (Unix seconds)
	Data      []Data // Array of sensor measurements
}

// MessageType returns the message type identifier for SensorCommand.
func (s *SensorCommand) MessageType() MsgType {
	return MsgTypeCommand
}

// MessageType returns the message type identifier for SensorConfig.
func (s *SensorConfig) MessageType() MsgType {
	return MsgTypeConfig
}

// MessageType returns the message type identifier for SensorHeartbeat.
func (s *SensorHeartbeat) MessageType() MsgType {
	return MsgTypeHeartbeat
}

// MessageType returns the message type identifier for SensorData.
func (s *SensorData) MessageType() MsgType {
	return MsgTypeSensorData
}

// MessageType returns the message type identifier for CustomData.
func (s *CustomData) MessageType() MsgType {
	return MsgTypeCustom
}

// MessageType returns the message type identifier for TimeSync.
func (t *TimeSync) MessageType() MsgType {
	return MsgTypeTimeSync
}

// MessageType returns the message type identifier for Ack.
func (a *Ack) MessageType() MsgType {
	return MsgTypeAck
}

// MessageType returns the message type identifier for Registration.
func (r *Registration) MessageType() MsgType {
	return MsgTypeRegister
}

// MessageType returns the message type identifier for Fragment.
func (f *Fragment) MessageType() MsgType {
	return MsgTypeFragment
}

// MessageType returns the message type identifier for RelayedMessage.
func (r *RelayedMessage) MessageType() MsgType {
	return MsgTypeRelayed
}

// MessageType returns the message type identifier for SensorDataMulti.
func (d *SensorDataMulti) MessageType() MsgType {
	return MsgTypeSensorDataMulti
}
