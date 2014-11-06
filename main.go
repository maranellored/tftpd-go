package main

import (
	"flag"
	"strconv"
)

var port, n_threads, timeout int
var address string

func init() {
	flag.IntVar(&port, "port", 7125, "The port to run the TFTP server on")
	flag.StringVar(&address, "address", "127.0.0.1", "The address on which to start the TFTP server")
	flag.IntVar(&n_threads, "threads", 16, "The max number of threads to service requests with")
	flag.IntVar(&timeout, "timeout", 1, "The timeout for the network connections in seconds")
}

func main() {
	flag.Parse()

	serverAddress := address + ":" + strconv.Itoa(port)

	startTFTPServer(serverAddress, n_threads, timeout)
}
