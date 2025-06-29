package message

type DataType int8
type Status uint8
type ConfigKey uint8
type CustomType uint8
type AckStatus uint8
type DeviceType uint8

const (
	Accelerometer DataType = 0x01
	Gyroscope     DataType = 0x02
	Quaternion    DataType = 0x03
	EulerAngles   DataType = 0x04
)

const (
	Ok          Status = 0x01
	Expectation Status = 0x02
	Collection  Status = 0x03
	LowBattery  Status = 0x04
	Error       Status = 0x05
)

const (
	ConfigKeySampleRate  ConfigKey = 0x01
	ConfigKeyRange       ConfigKey = 0x02
	ConfigKeyMAC         ConfigKey = 0x03
	ConfigKeyDeviceName  ConfigKey = 0x04
	ConfigKeyIPAddress   ConfigKey = 0x05
	ConfigKeyMode        ConfigKey = 0x06
	ConfigKeySensitivity ConfigKey = 0x07
	ConfigKeyCalibration ConfigKey = 0x08
)

const (
	CustomTypeLog    CustomType = 0x01
	CustomTypeError  CustomType = 0x02
	CustomTypeDebug  CustomType = 0x03
	CustomTypeString CustomType = 0x04
	CustomTypeBinary CustomType = 0x05
)

const (
	AckOK             AckStatus = 0x01
	AckError          AckStatus = 0x02
	AckInvalidCRC     AckStatus = 0x03
	AckUnknownMessage AckStatus = 0x04
	AckBufferFull     AckStatus = 0x05
)

const (
	DeviceTypeBNO085  DeviceType = 0x01
	DeviceTypeMPU6050 DeviceType = 0x02
	DeviceTypeLSM6DS3 DeviceType = 0x03
	DeviceTypeADXL345 DeviceType = 0x04
	DeviceTypeHub     DeviceType = 0x10
	DeviceTypeCustom  DeviceType = 0xFF
)

const (
	CapAccelerometer uint8 = 1 << 0
	CapGyroscope     uint8 = 1 << 1
	CapMagnetometer  uint8 = 1 << 2
	CapQuaternion    uint8 = 1 << 3
	CapTemperature   uint8 = 1 << 4
)

type Item struct {
	Key    ConfigKey
	Length uint8
	Value  []byte
}

type SensorCommand struct {
	SensorID  uint8
	TimeStamp uint32
	Command   uint8
}

type SensorConfig struct {
	SensorID  uint8
	TimeStamp uint32
	Config    []Item
}

type SensorHeartbeat struct {
	SensorID  uint8
	TimeStamp uint32
	Battery   uint8
	Status    Status
}

type SensorData struct {
	SensorID  uint8
	TimeStamp uint32
	Type      DataType
	Values    []float32
}

type CustomData struct {
	SensorID  uint8
	TimeStamp uint32
	DataType  CustomType
	Data      []Item
}

type TimeSync struct {
	SensorID   uint8
	ServerTime uint32
	SensorTime uint32
}

type Ack struct {
	SensorID  uint8
	MessageID uint16
	Status    AckStatus
}

type Registration struct {
	SensorID     uint8
	DeviceType   DeviceType
	Capabilities uint8
	FWVersion    uint16
}

type Fragment struct {
	MessageID      uint16
	FragmentNum    uint8
	TotalFragments uint8
	Data           []byte
}

func (s *SensorCommand) MessageType() MsgType {
	return MsgTypeCommand
}

func (s *SensorConfig) MessageType() MsgType {
	return MsgTypeConfig
}

func (s *SensorHeartbeat) MessageType() MsgType {
	return MsgTypeHeartbeat
}

func (s *SensorData) MessageType() MsgType {
	return MsgTypeSensorData
}

func (s *CustomData) MessageType() MsgType {
	return MsgTypeCustom
}

func (t *TimeSync) MessageType() MsgType {
	return MsgTypeTimeSync
}

func (a *Ack) MessageType() MsgType {
	return MsgTypeAck
}

func (r *Registration) MessageType() MsgType {
	return MsgTypeRegister
}

func (f *Fragment) MessageType() MsgType {
	return MsgTypeFragment
}
