package main

import (
	"bytes"
	"net"
	"testing"
)

type TestConnection struct {
	responseCache [][]byte
	requestCache  [][]byte
	addr          *net.UDPAddr
}

func (c *TestConnection) Write(packet []byte) (int, error) {
	c.responseCache[0] = packet
	return len(packet), nil
}

func (c *TestConnection) Read(buffer []byte) (int, *net.UDPAddr, error) {
	copy(buffer, c.requestCache[0])
	return len(c.requestCache[0]), c.addr, nil
}

func (c *TestConnection) GetRemoteAddr() *net.UDPAddr {
	return c.addr
}

func (c *TestConnection) GetConnection() *net.UDPConn {
	return nil
}

func TestProcessReadRequest(t *testing.T) {

	connection := &TestConnection{
		responseCache: make([][]byte, 512),
		requestCache:  make([][]byte, 512),
		addr: &net.UDPAddr{
			Port: 8088,
		},
	}

	fileCache := NewTftpCache()
	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}
	fileCache.putData("hello", data)

	ackPkt := CreateAckPacket(uint16(1))
	connection.requestCache[0] = ackPkt
	processReadRequest("hello", "octet", connection, fileCache)

	buffer := connection.responseCache[0]

	tmp, _, err := ParseDataPacket(buffer, len(buffer))
	if err != nil {
		t.Error("Couldnt parse the data packet. " + err.Error())
	}

	if !bytes.Equal(tmp, data) {
		t.Error("Bytes mangled during processing?")
	}

}

func TestProcessWriteRequest(t *testing.T) {

	connection := &TestConnection{
		responseCache: make([][]byte, 512),
		requestCache:  make([][]byte, 512),
		addr: &net.UDPAddr{
			Port: 8088,
		},
	}

	fileCache := NewTftpCache()
	data := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'}

	dataPkt := CreateDataPacket(uint16(1), data)
	connection.requestCache[0] = dataPkt

	processWriteRequest("hello", "octet", connection, fileCache)

	storedData := fileCache.getData("hello")

	if !bytes.Equal(data, storedData) {
		t.Error("Bytes mangled while being stored in file cache")
	}

}
