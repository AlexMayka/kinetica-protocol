# Kinetica Protocol

A universal, binary protocol for real-time sensor communication with multi-transport support. Designed for IoT devices, motion capture systems, and embedded sensor networks.

## 🚀 Features

- **Multi-Transport Support**: TCP, UDP, Serial/UART, and Bluetooth Low Energy (BLE)
- **Binary Protocol**: Efficient, compact message format with magic bytes validation
- **CRC Validation**: Transport-specific integrity checking (8-bit, 16-bit, 32-bit CRC)
- **Message Types**: 11 different message types for sensor data, commands, configuration, and more
- **Fragmentation Support**: Handle large messages across transport boundaries
- **Device Categories**: Support for 3-axis, 6-axis, 9-axis sensors and hub devices
- **Go Implementation**: Complete Go library with comprehensive test coverage

## 📋 Protocol Overview

### Packet Structure
```
┌──────────┬────────────┬──────────┬────────┬──────────┬──────────┐
│  Header  │  PacketID  │ Version  │  Type  │ Payload  │  Footer  │
│    2B    │     1B     │    1B    │   1B   │   0-XB   │   0-4B   │
└──────────┴────────────┴──────────┴────────┴──────────┴──────────┘
```

- **Magic Bytes**: `KN` (0x4B, 0x4E) for packet identification
- **PacketID**: Unique identifier (0-255, wraps around)
- **Version**: Protocol version (currently v1)
- **Message Type**: 11 supported message types
- **Variable Footer**: CRC validation based on transport type

### Message Types

| Type | ID | Description | Typical Size |
|------|----|-----------:|-------------|
| SensorCommand | 0x01 | Commands to sensors | 11B |
| SensorConfig | 0x02 | Configuration parameters | 20B+ |
| SensorHeartbeat | 0x03 | Status and battery info | 13B |
| SensorData | 0x04 | Single sensor measurement | 25B |
| CustomData | 0x05 | Application-specific data | Variable |
| TimeSync | 0x06 | Time synchronization | 15B |
| Ack | 0x07 | Acknowledgments | 10B |
| Registration | 0x08 | Sensor registration | 11B |
| Fragment | 0x09 | Message fragmentation | Variable |
| RelayedMessage | 0x0A | Hub/relay forwarding | 32B+ |
| SensorDataMulti | 0x0B | Multiple sensor readings | 77B+ |

### Device Types

- **3-Axis**: Accelerometer only
- **6-Axis**: Accelerometer + Gyroscope  
- **9-Axis**: Accelerometer + Gyroscope + Magnetometer
- **Hub**: Data concentrator/aggregator
- **Relay**: Message forwarding device
- **Custom**: Application-specific devices

## 🛠 Installation

```bash
go mod tidy
```

### Dependencies

- **BLE Transport**: `tinygo.org/x/bluetooth` for BLE communication
- **Serial Transport**: `go.bug.st/serial` for UART/RS232
- **Network**: Standard Go `net` package for TCP/UDP

## 🏃‍♂️ Quick Start

### TCP Client Example

```go
package main

import (
    "kinetica-protocol/protocol/message"
    "kinetica-protocol/transport/net"
    "time"
)

func main() {
    // Configure TCP transport
    config := net.Config{
        Address:      "localhost:8081",
        WriteTimeout: 5 * time.Second,
        ReadTimeout:  10 * time.Second,
    }

    // Create connection
    transport := net.NewTCP(config)
    conn, err := transport.Connection()
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Register sensor
    registration := &message.Registration{
        SensorID:     1,
        DeviceType:   message.DeviceType9Axis,
        Capabilities: message.CapAccelerometer | message.CapGyroscope,
        FWVersion:    0x0100,
    }

    err = conn.Send(registration, message.MsgTypeRegister)
    if err != nil {
        panic(err)
    }

    // Send sensor data
    sensorData := &message.SensorData{
        SensorID:  1,
        TimeStamp: uint32(time.Now().Unix()),
        Data: message.Data{
            Type:   message.Accelerometer,
            Values: []float32{0.1, 0.2, 9.8},
        },
    }

    err = conn.Send(sensorData, message.MsgTypeSensorData)
    if err != nil {
        panic(err)
    }
}
```

### BLE Client Example

```go
package main

import (
    "kinetica-protocol/transport/ble"
    "tinygo.org/x/bluetooth"
    "time"
)

func main() {
    config := ble.Config{
        DeviceAddress:  stringPtr("AA:BB:CC:DD:EE:FF"),
        ServiceUUID:    bluetooth.New16BitUUID(0x180F),
        WriteCharUUID:  bluetooth.New16BitUUID(0x2A19),
        NotifyCharUUID: bluetooth.New16BitUUID(0x2A1A),
        ScanTimeout:    10 * time.Second,
        ReadTimeout:    5 * time.Second,
    }

    transport := ble.NewBLE(config)
    conn, err := transport.Connection()
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Use connection for sensor communication...
}

func stringPtr(s string) *string { return &s }
```

## 🔧 Transport Comparison

| Transport | CRC Type | Max Message | Use Case | Reliability |
|-----------|----------|-------------|----------|-------------|
| **TCP** | None | 64KB | High-throughput servers | High |
| **UDP** | CRC8 | 1472B | Fast datagrams | Medium |
| **Serial** | CRC8 | 4KB | Embedded devices | High |
| **BLE** | CRC8 | 255B | IoT sensors | High |

### CRC Validation by Transport

- **TCP**: No CRC (relies on TCP reliability)
- **UDP/Serial/BLE**: 8-bit CRC for integrity checking
- **Configurable**: Support for 16-bit and 32-bit CRC when needed

## 📁 Project Structure

```
kinetica-protocol/
├── protocol/
│   ├── codec/          # Binary encoding/decoding
│   └── message/        # Message types and structures
├── transport/
│   ├── ble/           # Bluetooth Low Energy
│   ├── net/           # TCP and UDP
│   ├── serial/        # UART/RS232
│   └── *.go           # Transport interfaces
├── internal/
│   └── utils/         # CRC calculations
├── examples/          # Usage examples
└── docs/             # Protocol documentation
```

## 📡 Transport Details

### TCP Transport
- **Purpose**: High-throughput server applications
- **CRC**: None (TCP handles reliability)
- **Max Size**: 64KB
- **Features**: Client/server modes, connection pooling

### UDP Transport  
- **Purpose**: Fast, connectionless communication
- **CRC**: 8-bit for datagram validation
- **Max Size**: 1472 bytes (Ethernet MTU)
- **Features**: Single-socket server model

### Serial Transport
- **Purpose**: Embedded device communication
- **CRC**: 8-bit for line integrity
- **Max Size**: 4KB
- **Features**: Configurable baud rate, parity, flow control

### BLE Transport
- **Purpose**: IoT sensor connectivity
- **CRC**: 8-bit for wireless reliability
- **Max Size**: 255 bytes (BLE MTU)
- **Features**: GATT service discovery, notification handling

## 🧪 Testing

Run all tests:
```bash
go test ./...
```

Run transport-specific tests:
```bash
go test ./transport/net/
go test ./transport/ble/
go test ./transport/serial/
go test ./protocol/codec/
```

## 📚 Examples

The `examples/` directory contains complete working examples:

- `codec/` - Basic message encoding/decoding
- `tcp_client/` & `tcp_server/` - TCP client/server communication
- `udp_client/` & `udp_server/` - UDP datagram communication  
- `serial_client/` - Serial/UART device communication
- `ble_client/` - Bluetooth Low Energy sensor connection

Run an example:
```bash
cd examples/tcp_server
go run main.go
```

## 🤝 Architecture

### Message Flow
```
Application Layer
       ↕
Kinetica Protocol (codec)
       ↕
Transport Layer (TCP/UDP/Serial/BLE)
       ↕
Physical Layer (Network/UART/Bluetooth)
```

### Key Interfaces

- **`transport.Transport`**: Connection management
- **`transport.Connection`**: Message send/receive
- **`message.Message`**: Protocol message types
- **`codec.Marshal/Unmarshal`**: Binary encoding

## 📄 License

This project is available under the MIT License.

## 🔗 Related Projects

- **TinyGo Bluetooth**: BLE support for embedded devices
- **Go Serial**: Cross-platform serial port communication

---

Built for real-time sensor networks and IoT applications 🌐