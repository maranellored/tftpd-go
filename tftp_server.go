package main

import (
	"fmt"
	"net"
	"os"
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

func startTFTPServer(server string, threads, timeout int) {
	addr, err := net.ResolveUDPAddr("udp", server)
	handleError(err)

	conn, err := net.ListenUDP("udp", addr)
	handleError(err)

	fmt.Println("Started TFTP server at " + server)
	// Create a channel of size = # of threads
	tftpChannel := make(chan TftpRawRequest, threads)
	tftpCache := NewTftpCache()

	for i := 0; i < threads; i++ {
		go handleTftpRequest(tftpChannel, tftpCache, timeout)
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

	close(tftpChannel)
}

func handleTftpRequest(tftpChannel chan TftpRawRequest, fileCache *TftpCache, timeout int) {

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
			fmt.Println("Processing TFTP Write Request for file " + request.GetFilename() + ", in mode " + request.GetMode() + ", from host " + tftpConnection.GetRemoteAddr().String())
			processWriteRequest(request.GetFilename(), request.GetMode(), tftpConnection, fileCache)
		case RRQ:
			fmt.Println("Processing TFTP Read Request for file " + request.GetFilename() + ", in mode " + request.GetMode() + ", from host " + tftpConnection.GetRemoteAddr().String())
			processReadRequest(request.GetFilename(), request.GetMode(), tftpConnection, fileCache)
		default:
			fmt.Println("Received unknown request. Discarding..")
		}

		conn.Close()
	}
}

func processReadRequest(filename, mode string, conn Connection, fileCache *TftpCache) {

	if mode != "octet" {
		SendError(ERR_UNDEFINED, "Unsupported mode. Will only support octet mode", conn)
		return
	}

	array := fileCache.getData(filename)
	if array == nil {
		fmt.Println("File " + filename + " not found.")
		SendError(ERR_NOT_FOUND, "File not found", conn)
		return
	}

	var currentBlock uint16 = 1
	currentLength := 0

	for currentLength < len(array) {
		ackBuffer := make([]byte, 4)

		maxLength := currentLength + MAX_DATA_BLOCK_SIZE
		if maxLength > len(array) {
			maxLength = len(array)
		}
		dataPkt := CreateDataPacket(currentBlock, array[currentLength:maxLength])
		_, err := conn.Write(dataPkt)
		if err != nil {
			SendError(ERR_UNDEFINED, "Error while sending a data packet", conn)
			return
		}

		for {
			_, clientAddr, err := conn.Read(ackBuffer)
			if err != nil {
				SendError(ERR_UNDEFINED, "Error reading ACK packet off the wire", conn)
			}
			if clientAddr.Port != conn.GetRemoteAddr().Port {
				// This packet came in from an unknown host.
				// Reuse our connection object but set remote addr to the erroneous client.
				SendError(ERR_UNKNOWN_TID, "Error packet from uknown source", TftpConnection{conn.GetConnection(), clientAddr})
				continue
			}
			break
		}

		ackBlockNumber, err := ParseAckPacket(ackBuffer)
		if err != nil {
			SendError(ERR_ILLEGAL_TFTP_OP, "Illegal packet sent. Expected ACK packet", conn)
			return
		}

		// We dont really check for out of order block numbers here
		// because the client will send us its required block number when it receives
		// an out of order packet.
		currentLength = int(currentBlock) * MAX_DATA_BLOCK_SIZE
		currentBlock = ackBlockNumber + 1
	}

	fmt.Println("TFTP READ: File sent back to requester- " + filename)

}

func processWriteRequest(filename, mode string, conn Connection, fileCache *TftpCache) {
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
		ackPkt := CreateAckPacket(currentBlock)
		_, err := conn.Write(ackPkt)
		if err != nil {
			SendError(ERR_UNDEFINED, "Error while sending a data packet", conn)
			return
		}

		if bytesRead < MAX_PACKET_SIZE {
			break
		}

		dataPktBuffer := make([]byte, MAX_PACKET_SIZE)
		var n int
		var clientAddr *net.UDPAddr

		for {
			n, clientAddr, err = conn.Read(dataPktBuffer)
			if err != nil {
				SendError(ERR_UNDEFINED, err.Error(), conn)
				return
			}

			if clientAddr.Port != conn.GetRemoteAddr().Port {
				// This packet came in from an unknown host.
				// Reuse our connection object but set remote addr to the erroneous client.
				SendError(ERR_UNKNOWN_TID, "Error packet from unknown source", TftpConnection{conn.GetConnection(), clientAddr})
				continue
			}
			break
		}

		data, block, err := ParseDataPacket(dataPktBuffer, n)
		if err != nil {
			SendError(ERR_ILLEGAL_TFTP_OP, "Illegal packet sent. Expected data packet.", conn)
			return
		}

		// if we've already seen this data packet or receive an out of order
		// data pkt, just continue and re-ack the previous block that we've seen
		if block-1 != currentBlock {
			continue
		}

		// The size of the data received is
		tmpDataArray = append(tmpDataArray, data...)
		bytesRead = n
		currentBlock = block
	}

	fileCache.putData(filename, tmpDataArray)
	fmt.Println("TFTP WRITE: File written to memory- " + filename)
}

func handleError(err error) {
	if err != nil {
		fmt.Println("ERROR: " + err.Error())
		os.Exit(1)
	}
}
