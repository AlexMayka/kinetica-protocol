# Kinetica Protocol Specification

## Protocol Structure

### Packet Format

```
┌──────────────────────────────────────────────────────────────────┐
│                           Kinetica Packet                        │
├──────────┬────────────┬──────────┬────────┬──────────┬──────────┤
│  Header  │  PacketID  │ MsgType  │  Len   │ Payload  │  Footer  │
│    2B    │     1B     │    1B    │   1B   │   0-XB   │   1-4B   │
└──────────┴────────────┴──────────┴────────┴──────────┴──────────┘
```

### Header Structure (6 bytes total)

```
┌─────────────┬────────────┬──────────┬────────┐
│ Magic Bytes │  PacketID  │ MsgType  │  Len   │
│  "KN" (2B)  │     1B     │    1B    │   1B   │
└─────────────┴────────────┴──────────┴────────┘
```

### Footer Structure (Variable size based on CRC type)

```
Transport CRC Types:
┌──────────────────┬──────────┬─────────────┐
│ TransportNone    │   0B     │ No footer   │
│ TransportLength  │   1B     │ Length only │  
│ TransportCRC8    │   1B     │ CRC8        │
│ TransportCRC16   │   2B     │ CRC16       │
│ TransportCRC32   │   4B     │ CRC32       │
└──────────────────┴──────────┴─────────────┘
```

## Message Types and Sizes

### Core Messages

| Message Type | ID | Min Size | Typical Size | Max Size | Description |
|--------------|----|---------:|-------------:|---------:|-------------|
| SensorCommand | 0x01 | 11B | 11B | 11B | Command to sensor |
| SensorConfig | 0x02 | 12B | 20B | ~1KB | Sensor configuration |
| SensorHeartbeat | 0x03 | 13B | 13B | 13B | Heartbeat with battery |
| SensorData | 0x04 | 25B | 25B | 25B | Single axis data |
| CustomData | 0x05 | 12B | 50B | ~64KB | Custom payload |
| TimeSync | 0x06 | 15B | 15B | 15B | Time synchronization |
| Ack | 0x07 | 10B | 10B | 10B | Acknowledgment |
| Registration | 0x08 | 11B | 11B | 11B | Sensor registration |
| Fragment | 0x09 | 12B | 512B | ~64KB | Message fragmentation |

### Extended Messages

| Message Type | ID | Min Size | Typical Size | Max Size | Description |
|--------------|----|---------:|-------------:|---------:|-------------|
| RelayedMessage | 0x0A | 10B | 34B | ~64KB | Relayed through ESP-NOW |
| SensorDataMulti | 0x0B | 15B | 77B | 77B | Multiple sensor data (planned) |

## Device Types

```
Classification by Sensor Axes:
┌─────────────────┬──────┬─────────────────────────────────┐
│ Device Type     │  ID  │ Description                     │
├─────────────────┼──────┼─────────────────────────────────┤
│ DeviceType3Axis │ 0x01 │ 3-axis (Accelerometer only)    │
│ DeviceType6Axis │ 0x02 │ 6-axis (Accel + Gyroscope)     │
│ DeviceType9Axis │ 0x03 │ 9-axis (Accel + Gyro + Mag)    │
│ DeviceTypeHub   │ 0x10 │ Data concentrator/hub           │
│ DeviceTypeRelay │ 0x11 │ ESP-NOW relay/retransmitter     │
│ DeviceTypeCustom│ 0xFF │ Custom/user-defined types       │
└─────────────────┴──────┴─────────────────────────────────┘
```

## Capabilities Flags

```
Sensor Capabilities (Bitfield):
┌──────────────────┬─────────┬──────────────────────────────┐
│ Capability       │ Bit Pos │ Description                  │
├──────────────────┼─────────┼──────────────────────────────┤
│ CapAccelerometer │    0    │ Has accelerometer (0x01)     │
│ CapGyroscope     │    1    │ Has gyroscope (0x02)         │
│ CapMagnetometer  │    2    │ Has magnetometer (0x04)      │
│ CapQuaternion    │    3    │ Can output quaternions (0x08)│
│ CapTemperature   │    4    │ Has temperature sensor (0x10)│
└──────────────────┴─────────┴──────────────────────────────┘
```

## Transport Support Matrix

```
┌─────────────┬─────┬─────┬─────────┬────────┬─────────┐
│ Transport   │ TCP │ UDP │ SERIAL  │  BLE   │ ESP-NOW │
├─────────────┼─────┼─────┼─────────┼────────┼─────────┤
│ Status      │  ✅  │ 🔄  │   🔄    │   🔄   │   🔄    │
│ Max Packet  │ 64K │ 64K │  Unlimited │ ~240B  │  250B   │
│ Reliability │ High│ Low │   High  │  High  │   Low   │
│ Use Case    │Demo │Fast │ESP-NOW  │Config  │ Mesh    │
└─────────────┴─────┴─────┴─────────┴────────┴─────────┘

Legend: ✅ Implemented, 🔄 Planned
```

## Packet Size Examples

### Minimum Packets (with TransportNone)
```
Registration:     [KN][01][08][05][sensor_data]           = 11 bytes
Ack:             [KN][01][07][04][ack_data]               = 10 bytes  
Heartbeat:       [KN][01][03][07][heartbeat_data]         = 13 bytes
SensorData:      [KN][01][04][13][sensor_data]            = 25 bytes
```

### Maximum Practical Packets
```
SensorConfig:    [KN][01][02][~1000][config_items][CRC32] = ~1KB
CustomData:      [KN][01][05][~64K][custom_payload][CRC32] = ~64KB  
Fragment:        [KN][01][09][~64K][fragment_data][CRC32]  = ~64KB
RelayedMessage:  [KN][01][0A][32][relay_id + orig_packet] = varies
```

### Transport Overhead
```
TransportNone:    +0 bytes (no footer)
TransportLength:  +1 byte  (length check)
TransportCRC8:    +1 byte  (basic CRC)
TransportCRC16:   +2 bytes (standard CRC)
TransportCRC32:   +4 bytes (robust CRC)
```

## System Architecture

### Transport Layer Overview
```
┌─────────────────────────────────────────────────────────────────┐
│                        Application Layer                        │
├─────────────────────────────────────────────────────────────────┤
│                       Kinetica Protocol                         │
├─────────────┬─────────────┬─────────────┬─────────────┬─────────┤
│ TCP Transport│ UDP Transport│ BLE Transport│Serial Transport│ ... │
├─────────────┼─────────────┼─────────────┼─────────────┼─────────┤
│    TCP/IP   │   UDP/IP    │  Bluetooth  │    UART     │   ...   │
└─────────────┴─────────────┴─────────────┴─────────────┴─────────┘
```

### Data Flow Architecture
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Sensor    │───▶│    ESP32    │───▶│   Gateway   │───▶│   Server    │
│  (BNO085)   │    │  (ESP-NOW)  │    │  (Serial)   │    │   (TCP)     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
      │                    │                    │                    │
    BLE Config        ESP-NOW Mesh         Serial Link         Network
   (25-100B/msg)      (25-32B/msg)        (32B relayed)      (Aggregated)
```

## Communication Flows

### Sensor Registration Flow
```
Sensor                           Server
  │                                │
  │─────── Registration ─────────▶│ (11B)
  │        [SensorID, Type,       │
  │         Capabilities]         │
  │                               │
  │◀────────── Ack ──────────────│ (10B)
  │          [Status: OK]         │
  │                               │
  │── Ready for data exchange ───│
```

### Data Transmission Flow  
```
Sensor                           Server
  │                                │
  │─────── SensorData ──────────▶│ (25B)
  │       [Accel X,Y,Z values]    │
  │                               │
  │─────── Heartbeat ───────────▶│ (13B)
  │       [Battery: 85%]          │
  │                               │
  │◀────── TimeSync ─────────────│ (15B)
  │      [Server timestamp]       │
```

### ESP-NOW Relay Flow
```
Sensor ───▶ ESP32 Relay ───▶ Gateway ───▶ Server
   │            │               │            │
   │    25B     │     32B       │    32B     │   Processed
   │            │               │            │
   │  Original  │   Relayed     │  Relayed   │   Data
   │  Packet    │   Packet      │  Packet    │
   │            │               │            │
┌─────────┐ ┌─────────────────┐ ┌──────────┐ ┌──────────┐
│ [Data]  │ │[RelayID][Data]  │ │[Relayed] │ │Extract & │
│         │ │                 │ │Message   │ │Process   │
└─────────┘ └─────────────────┘ └──────────┘ └──────────┘
```

### Multi-Transport Configuration
```
                           ┌─────────────┐
                           │   Server    │
                           │  (TCP/UDP)  │
                           └──────┬──────┘
                                  │
                            ┌─────┴─────┐
                            │  Gateway  │
                            │ (Serial)  │
                            └─────┬─────┘
                                  │
                      ┌───────────┴───────────┐
                      │    ESP32 Relay       │
                      │    (ESP-NOW Hub)     │
                      └───────────┬───────────┘
                                  │
            ┌─────────────────────┼─────────────────────┐
            │                     │                     │
    ┌───────┴───────┐     ┌───────┴───────┐     ┌───────┴───────┐
    │   Sensor 1    │     │   Sensor 2    │     │   Sensor 3    │
    │ (BLE Config)  │     │ (BLE Config)  │     │ (BLE Config)  │
    │ (ESP-NOW Data)│     │ (ESP-NOW Data)│     │ (ESP-NOW Data)│
    └───────────────┘     └───────────────┘     └───────────────┘
```

### Message Size Progression
```
Original Sensor Data:
┌─────────────────────────────────────┐
│ [KN][01][04][13][SensorData] = 25B  │
└─────────────────────────────────────┘

After ESP-NOW Relay:
┌──────────────────────────────────────────────────────────┐  
│ [KN][02][0A][1A][RelayID][Original_25B_Packet] = 32B    │
└──────────────────────────────────────────────────────────┘

After Multi-Axis Aggregation:
┌─────────────────────────────────────────────────────────────────────┐
│ [KN][03][0B][47][DataMask][Accel][Gyro][Mag][Quat][Euler] = 77B    │
└─────────────────────────────────────────────────────────────────────┘
```

## Usage Patterns

### BLE Configuration Sequence
```
Mobile App                     Sensor (BLE Peripheral)
    │                               │
    │─────── Scan & Connect ──────▶│
    │                               │
    │─────── Read Capabilities ───▶│
    │◀────── [9-Axis, BNO085] ─────│
    │                               │
    │─────── SensorConfig ────────▶│ (Sample rate, range, etc.)
    │◀────── Ack [OK] ─────────────│
    │                               │
    │─────── Disconnect ───────────│ (Switch to ESP-NOW mesh)
```

### ESP-NOW Mesh Operation
```
Time: 0ms     100ms    200ms    300ms    400ms
  │           │        │        │        │
Sensor1 ─────▶Data ────────────▶HB ──────────▶Data
Sensor2 ──────────▶Data ────────────▶HB ──────────▶Data  
Sensor3 ────────────────▶Data ────────────▶HB ──────────▶
  │           │        │        │        │
Relay   ─────▶Fwd ────▶Fwd ────▶Fwd ────▶Fwd
  │           │        │        │        │
Gateway ─────▶Recv────▶Recv────▶Recv────▶Recv
  │           │        │        │        │
Server  ─────▶Proc────▶Proc────▶Proc────▶Proc

Legend: Data=SensorData(25B), HB=Heartbeat(13B), Fwd=RelayedMessage(32B)
```