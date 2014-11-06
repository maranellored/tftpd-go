package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
)

type Connection interface {
	Write(packet []byte) (int, error)
	Read(buffer []byte) (int, *net.UDPAddr, error)
	GetRemoteAddr() *net.UDPAddr
	GetConnection() *net.UDPConn
}

type TftpConnection struct {
	Connection *net.UDPConn
	RemoteAddr *net.UDPAddr
}

func (c TftpConnection) Write(packet []byte) (int, error) {
	return c.Connection.WriteToUDP(packet, c.RemoteAddr)
}

func (c TftpConnection) Read(buffer []byte) (int, *net.UDPAddr, error) {
	return c.Connection.ReadFromUDP(buffer)
}

func (c TftpConnection) GetRemoteAddr() *net.UDPAddr {
	return c.RemoteAddr
}

func (c TftpConnection) GetConnection() *net.UDPConn {
	return c.Connection
}

// The raw request is a struct that holds data about the client
// and the request as taken from the wire.
type TftpRawRequest struct {
	Addr     *net.UDPAddr
	RawBytes []byte
}

func (req TftpRawRequest) GetAddr() *net.UDPAddr {
	return req.Addr
}

func (req TftpRawRequest) GetRawBytes() []byte {
	return req.RawBytes
}

type TftpRequest struct {
	Opcode   uint16
	Filename string
	Mode     string
}

func (r TftpRequest) GetOpcode() uint16 {
	return r.Opcode
}

func (r TftpRequest) GetFilename() string {
	return r.Filename
}

func (r TftpRequest) GetMode() string {
	return r.Mode
}

func ParseAckPacket(buffer []byte) (uint16, error) {
	opcode := binary.BigEndian.Uint16(buffer[:2])

	if opcode != ACK {
		return 0, errors.New("Expected ACK packet but received something else")
	}

	ackBlock := binary.BigEndian.Uint16(buffer[2:4])
	return ackBlock, nil
}

func ParseDataPacket(buffer []byte, bytesRead int) ([]byte, uint16, error) {
	opcode := binary.BigEndian.Uint16(buffer[:2])
	if opcode != DATA {
		return nil, 0, errors.New("Expected data packet but received something else")
	}

	blockNumber := binary.BigEndian.Uint16(buffer[2:4])

	// Only return the relevant bytes
	return buffer[4:bytesRead], blockNumber, nil
}

func CreateErrorPacket(errCode uint16, errMsg string) []byte {

	buffer := new(bytes.Buffer)
	// Write the opcode first
	binary.Write(buffer, binary.BigEndian, ERROR)
	binary.Write(buffer, binary.BigEndian, errCode)

	buffer.WriteString(errMsg)

	return buffer.Bytes()
}

func CreateAckPacket(blockNumber uint16) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, ACK)
	binary.Write(buffer, binary.BigEndian, blockNumber)

	return buffer.Bytes()
}

func CreateDataPacket(blockNumber uint16, data []byte) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, DATA)
	binary.Write(buffer, binary.BigEndian, blockNumber)

	buffer.Write(data)

	return buffer.Bytes()
}

func SendError(errorCode uint16, errorMsg string, connection Connection) {
	errorPkt := CreateErrorPacket(errorCode, errorMsg)
	connection.Write(errorPkt)
}

func ParseTftpRequest(buffer []byte) (TftpRequest, error) {
	opcode := binary.BigEndian.Uint16(buffer[:2])

	// Get the index of the first occurance of the null character
	n := bytes.Index(buffer[2:], []byte{0})
	if n-1 < 1 {
		// Filename length is 0. Throw error
		return TftpRequest{}, errors.New("Invalid filename specified")
	}
	filename := string(buffer[2 : n+2])

	// The nth byte here is the null character.
	// Move ahead by 1 and determine the mode
	m := bytes.Index(buffer[n+3:], []byte{0})
	if m-1 < 1 {
		// Mode length is 0. Throw error
		return TftpRequest{}, errors.New("Invalid mode specified")
	}
	mode := string(buffer[n+3 : n+3+m])

	return TftpRequest{opcode, filename, mode}, nil
}
