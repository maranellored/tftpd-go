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
