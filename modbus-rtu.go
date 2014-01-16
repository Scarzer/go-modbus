// Package modbusclient provides modbus Serial Line/RTU and TCP/IP access
// for client (master) applications to communicate with server (slave)
// devices. Logic specifically in this file implements the Serial Line/RTU
// protocol.

package modbusclient

import (
	//"errors"
	//"fmt"
	"os"
	"time"
)

// crc computes and returns a cyclic redundancy check of the given byte array
func crc(data []byte) uint16 {
	var crc16 uint16 = 0xffff
	l := len(data)
	for i := 0; i < l; i++ {
		crc16 ^= uint16(data[i])
		for j := 0; j < 8; j++ {
			if crc16&0x0001 > 0 {
				crc16 = (crc16 >> 1) ^ 0xA001
			} else {
				crc16 >>= 1
			}
		}
	}
	return crc16
}

// GenerateRTUFrame is a method corresponding to a RTUFrame object which returns a byte array representing the associated serial line/RTU application data unit (ADU)
func (frame *RTUFrame) GenerateRTUFrame() []byte {

	packetLen := 8
	if len(frame.Data) > 0 {
		packetLen = RTU_FRAME_MAXSIZE
	}

	bytesUsed := 0
	packet := make([]byte, packetLen)
	packet[0] = frame.SlaveAddress
	packet[1] = frame.FunctionCode
	packet[2] = byte(frame.StartRegister >> 8)       // (High Byte)
	packet[3] = byte(frame.StartRegister & 0xff)     // (Low Byte)
	packet[4] = byte(frame.NumberOfRegisters >> 8)   // (High Byte)
	packet[5] = byte(frame.NumberOfRegisters & 0xff) // (Low Byte)
	bytesUsed += 6

	if len(frame.Data) > 0 {
		nr := int(frame.NumberOfRegisters) * 2
		packet[bytesUsed] = byte(nr)
		bytesUsed += 1

		for i := 0; i < nr; i++ {
			packet[(bytesUsed + i)] = frame.Data[i]
		}
		bytesUsed += nr
	}

	// add the crc to the end
	packet_crc := crc(packet[:bytesUsed])
	packet[bytesUsed] = byte(packet_crc & 0xff)
	packet[(bytesUsed + 1)] = byte(packet_crc >> 8)

	return packet
}

func (frame *RTUFrame) Read(fd *os.File) ([]byte, error) {
	// generate the ADU from the RTU frame
	adu := frame.GenerateRTUFrame()

	// transmit the ADU to the slave device via the
	// serial port represented by the fd pointer
	_, err := fd.Write(adu)
	if err != nil {
		return []byte{}, MODBUS_EXCEPTIONS[EXCEPTION_GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND]
	}

	// allow the slave device adequate time to repond
	time.Sleep(300 * time.Millisecond)

	// then attempt to read the reply
	response := make([]byte, TCP_FRAME_MAXSIZE)
	_, err = fd.Read(response)
	if err != nil {
		return []byte{}, MODBUS_EXCEPTIONS[EXCEPTION_GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND]
	}

	// check the validity of the response

	return response, nil
}

/*
// RTURead
//   parameters
//   int        fd:  file descripter for serial device
//   byte  addr:  slave device address
//   byte  code:  function code
//   uint16 sr:    starting register number
//   uint16 nr:    number of registers to read
//   byte data[]: memory area for read data
func RTURead(fd *os.File, addr, code byte, sr, nr uint16) ([]byte, error) {
	//Preparation for Sending a Packet
	var send_packet = make([]byte, 8)

	//Packet Construction
	send_packet[0] = addr            // Slave Address
	send_packet[1] = code            // Function Code 0x03 = Multiple Read
	send_packet[2] = byte(sr >> 8)   // Start Register (High Byte)
	send_packet[3] = byte(sr & 0xff) // Start Register (Low Byte)
	send_packet[4] = byte(nr >> 8)   // Number of Registers (High Byte)
	send_packet[5] = byte(nr & 0xff) // Number of Registers (Low Byte)

	//Add CRC16
	send_packet_crc := crc(send_packet[:6])
	send_packet[6] = byte(send_packet_crc & 0xff)
	send_packet[7] = byte(send_packet_crc >> 8)

	// Preparation for Receiving a Packet
	var recv_packet = make([]byte, 256)
	_, err := fd.Write(send_packet)
	if err != nil {
		return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION")
	}
	time.Sleep(300 * time.Millisecond)
	_, err = fd.Read(recv_packet)
	if err != nil {
		return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION")
	}

	// Parse the Response
	if recv_packet[0] != send_packet[0] || recv_packet[1] != send_packet[1] {
		if recv_packet[0] == send_packet[0] && recv_packet[1]&0x7f == send_packet[1] {
			switch recv_packet[2] {
			case 1:
				return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_FUNCTION")
			case 2:
				return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_ADDRESS")
			case 3:
				return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_VALUE")
			case 4:
				return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_OPERATION")
			}
		}
		return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION")
	}

	//CRC check
	l := recv_packet[2]
	recv_packet_crc := crc(recv_packet[:3+l])
	if recv_packet[3+l] != byte((recv_packet_crc&0xff)) || recv_packet[3+l+1] != byte((recv_packet_crc>>8)) {
		return []byte{}, errors.New("MODBUS_ERROR_COMMUNICATION")
	}
	return recv_packet[3 : l+3], nil
}*/

/*
// RTUWrite
//   parameters
//   int        fd:  file descripter for serial device
//   byte  addr:  slave device address
//   byte  code:  function code
//   uint16 sr:    starting register number
//   uint16 nr:    number of registers to write
//   byte data[]: memory area for writing data
func RTUWrite(fd *os.File, addr, code byte, sr, nr uint16, data []byte) error {
	var send_packet = make([]byte, 256)

	// Packet Construction
	send_packet[0] = addr            // Slave Address
	send_packet[1] = code            // Function Code 0x10 = Multiple Write
	send_packet[2] = byte(sr >> 8)   // Start Register (High Byte)
	send_packet[3] = byte(sr & 0xff) // Start Register (Low Byte)
	send_packet[4] = byte(nr >> 8)   // Number of Registers (High Byte)
	send_packet[5] = byte(nr & 0xff) // Number of Registers (Low Byte)
	send_packet[6] = byte(nr * 2)

	for i := 0; i < int((nr * 2)); i++ {
		send_packet[7+i] = data[i]
	}

	length := 7 + nr*2 + 2
	// Add CRC16
	send_packet_crc := crc(send_packet[:length-2])
	send_packet[length-2] = byte(send_packet_crc & 0xff)
	send_packet[length-1] = byte(send_packet_crc >> 8)

	// Preparation for Receiving a Packet
	var recv_packet = make([]byte, 256)
	_, err := fd.Write(send_packet)
	if err != nil {
		return errors.New("MODBUS_ERROR_COMMUNICATION")
	}
	time.Sleep(300 * time.Millisecond)
	_, err = fd.Read(recv_packet)
	if err != nil {
		return errors.New("MODBUS_ERROR_COMMUNICATION")
	}

	// Parse the Response
	if recv_packet[0] != send_packet[0] || recv_packet[1] != send_packet[1] {
		if recv_packet[0] == send_packet[0] && recv_packet[1]&0x7f == send_packet[1] {
			switch recv_packet[2] {
			case 1:
				return errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_FUNCTION")
			case 2:
				return errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_ADDRESS")
			case 3:
				return errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_VALUE")
			case 4:
				return errors.New("MODBUS_ERROR_COMMUNICATION_ILLEGAL_OPERATION")
			}
		}
		return errors.New("MODBUS_ERROR_COMMUNICATION")
	}

	//Target Data Filed Check
	if recv_packet[2] == send_packet[2] && recv_packet[3] == send_packet[3] && recv_packet[4] == send_packet[4] && recv_packet[5] == send_packet[5] {
		//CRC check
		recv_packet_crc := crc(recv_packet[:6])
		if recv_packet[6] == byte((recv_packet_crc&0xff)) && recv_packet[7] == byte((recv_packet_crc>>8)) {
			return nil
		}
	}
	return errors.New("MODBUS_ERROR_COMMUNICATION")
}*/
