package main

import "testing"

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
