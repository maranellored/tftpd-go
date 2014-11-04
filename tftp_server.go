package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// Each opcode is 2 bytes. Use the unsigned 16bit int
const (
	_            = iota
	RRQ   uint16 = iota
	WRQ   uint16 = iota
	DATA  uint16 = iota
	ACK   uint16 = iota
	ERROR uint16 = iota
)

// Error codes
const (
	ERR_UNDEFINED        uint16 = iota
	ERR_NOT_FOUND        uint16 = iota
	ERR_ACCESS_VIOLATION uint16 = iota
	ERR_DISK_FULL        uint16 = iota
	ERR_ILLEGAL_TFTP_OP  uint16 = iota
	ERR_UNKNOWN_TID      uint16 = iota
	ERR_FILE_EXISTS      uint16 = iota
	ERR_NO_SUCH_USER     uint16 = iota
)

const MAX_PACKET_SIZE = 516
const MAX_DATA_BLOCK_SIZE = 512

var storedFiles = map[string][]byte{}

func startTFTPServer(server string, threads, timeout int) {
	addr, err := net.ResolveUDPAddr("udp", server)
	handleError(err)

	conn, err := net.ListenUDP("udp", addr)
	handleError(err)

	tftpChannel := make(chan TftpRawRequest)

	for i := 0; i < threads; i++ {
		go handleTftpRequest(tftpChannel, timeout)
	}

	for {
		// Dont bother with error checking here
		buffer := make([]byte, MAX_PACKET_SIZE)
		n, remoteAddr, _ := conn.ReadFromUDP(buffer)

		request := TftpRawRequest{
			Addr:     remoteAddr,
			RawBytes: buffer[:n],
		}

		tftpChannel <- request
	}
}

func handleTftpRequest(tftpChannel chan TftpRawRequest, timeout int) {

	for {
		rawRequest := <-tftpChannel
		// Create a new connection from localhost to this client
		localAddr, err := net.ResolveUDPAddr("udp", ":0")
		if err != nil {
			fmt.Println("Error opening new UDP listener. Discarding request")
			continue
		}

		conn, err := net.ListenUDP("udp", localAddr)
		if err != nil {
			fmt.Println("Error opening new UDP listener. Discarding request")
			continue
		}
		conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))
		//fmt.Println("Listening on:  " + conn.LocalAddr().String())
		rawRequestBuffer := rawRequest.GetRawBytes()

		tftpConnection := TftpConnection{
			Connection: conn,
			RemoteAddr: rawRequest.GetAddr(),
		}
		// Get the first 2 bytes of the request to read the op-code
		request, err := ParseTftpRequest(rawRequestBuffer)
		if err != nil {
			SendError(ERR_ILLEGAL_TFTP_OP, err.Error(), tftpConnection)
			conn.Close()
			continue
		}

		switch request.GetOpcode() {
		case WRQ:
			processWriteRequest(request.GetFilename(), request.GetMode(), tftpConnection)
		case RRQ:
			processReadRequest(request.GetFilename(), request.GetMode(), tftpConnection)
		default:
			fmt.Println("Received unknown request. Discarding..")
		}

		conn.Close()
	}
}

func processReadRequest(filename, mode string, conn TftpConnection) {

	if mode != "octet" {
		SendError(ERR_UNDEFINED, "Unsupported mode. Will only support octet mode", conn)
		return
	}

	array := storedFiles[filename]
	if array == nil {
		SendError(ERR_NOT_FOUND, "File not found", conn)
		return
	}

	var currentBlock uint16 = 1
	currentLength := 0

	for currentLength < len(array) {
		ackBuffer := make([]byte, MAX_PACKET_SIZE)

		maxLength := currentLength + MAX_DATA_BLOCK_SIZE
		if maxLength > len(array) {
			maxLength = len(array)
		}
		dataPkt := CreateDataPacket(currentBlock, array[currentLength:maxLength])
		conn.Write(dataPkt)

		conn.Read(ackBuffer)
		// TODO: Check the remoteaddr returned above to make sure we're talking to the right
		// Client
		ackBlockNumber, err := ParseAckPacket(ackBuffer)
		if err != nil {
			SendError(ERR_ILLEGAL_TFTP_OP, "Illegal packet sent", conn)
			return
		}

		// TODO: Check to make sure that we dont get duplicate ACKs for a lower
		// block number than that we are already at.
		// If we are 6 and ack comes for 4, ignore that.
		currentLength = int(currentBlock) * MAX_DATA_BLOCK_SIZE
		currentBlock = ackBlockNumber + 1
	}

}

func processWriteRequest(filename, mode string, conn TftpConnection) {
	if mode != "octet" {
		SendError(ERR_UNDEFINED, "Unsupported mode. Will only support octet mode", conn)
		return
	}

	// Create an initial array of size 512. We'll append to this
	// array if we need to keep adding more.
	var tmpDataArray []byte
	currentBlock := uint16(0)
	bytesRead := MAX_PACKET_SIZE

	for {
		// Send the first ack with block number of 0
		fmt.Println("Sending ACK packet with ACK: " + strconv.Itoa(int(currentBlock)))
		ackPkt := CreateAckPacket(currentBlock)
		conn.Write(ackPkt)

		if bytesRead < MAX_PACKET_SIZE {
			break
		}

		dataPktBuffer := make([]byte, MAX_PACKET_SIZE)
		n, _, err := conn.Read(dataPktBuffer)
		if err != nil {
			SendError(ERR_UNDEFINED, err.Error(), conn)
			return
		}

		// Make sure block numbers are increasing...
		data, block, err := ParseDataPacket(dataPktBuffer, n)
		if err != nil {
			SendError(ERR_ILLEGAL_TFTP_OP, "Illegal packet sent. Expected data packet.", conn)
			return
		}

		// The size of the data received is
		tmpDataArray = append(tmpDataArray, data...)
		bytesRead = n
		fmt.Println("Bytes read: " + strconv.Itoa(n))
		currentBlock = block
	}

	storedFiles[filename] = tmpDataArray
	fmt.Println("Size of data: " + strconv.Itoa(len(tmpDataArray)))
}

func handleError(err error) {
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		os.Exit(1)
	}
}
