package main

import (
	"strconv"
	"testing"
)

func TestParseAckPacket(t *testing.T) {
	buffer := []byte{0, 4, 0, 4}

	blockNumber, err := ParseAckPacket(buffer)
	if err != nil {
		t.Error("Did not expect an error")
	}

	if blockNumber != 4 {
		t.Error("Expected block number of 4")
	}
}

func TestParseAckPacketFail(t *testing.T) {
	buffer := []byte{0, 8, 0, 4}

	_, err := ParseAckPacket(buffer)
	if err == nil {
		t.Error("Expected an error while parsing Ack packet")
	}
}

func TestParseDataPacket(t *testing.T) {
	buffer := []byte{0, 3, 0, 10, 't', 'h', 'i', 's', 'd', 'a', 't', 'a'}

	data, blkNumber, err := ParseDataPacket(buffer, len(buffer))
	if err != nil {
		t.Error("Error parsing data packet: " + err.Error())
	}

	if blkNumber != 10 {
		t.Error("Error parsing block number. Got:  " + strconv.Itoa(int(blkNumber)))
	}

	if string(data) != "thisdata" {
		t.Error("Error parsing data. Got: " + string(data))
	}
}

func TestParseDataPacketFail(t *testing.T) {
	buffer := []byte{0, 9, 0, 10, 't', 'h', 'i', 's', 'd', 'a', 't', 'a'}

	_, _, err := ParseDataPacket(buffer, len(buffer))
	if err == nil {
		t.Error("Expected error while parsing data packet! ")
	}
}

func TestCreateAckPacket(t *testing.T) {
	blockNumber := uint16(9)

	pkt := CreateAckPacket(blockNumber)

	number, err := ParseAckPacket(pkt)
	if err != nil {
		t.Error("Error parsing ACK packet. " + err.Error())
		t.Error("Did not expect an error while parsing the ACK packet")
	}

	if blockNumber != number {
		t.Error("Expected the block number after creation to be same as one before")
	}
}

func TestCreateDataPacket(t *testing.T) {
	blockNumber := uint16(10)
	data := []byte{'t', 'h', 'i', 's', 'i', 's', 'd', 'a', 't', 'a'}

	pkt := CreateDataPacket(blockNumber, data)

	d, number, err := ParseDataPacket(pkt, len(pkt))
	if err != nil {
		t.Error("Error parsing data packet. " + err.Error())
	}

	if blockNumber != number {
		t.Error("Expected the block number after creation to be same as one before")
	}

	if string(d) != string(data) {
		t.Error("Mismatch in data after parsing. Expected: " + string(data) + " . But got: " + string(d))
	}

}
