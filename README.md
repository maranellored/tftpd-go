TFTP
----

A simple tftp server in Go.
The server is implemented using [RFC 1350](https://www.ietf.org/rfc/rfc1350.txt)

Usage
-----
Install go on your machine by following instructions at [Golang official website](http://golang.org)

Once installed, run the following commands to get the code and run it

```
$ go get github.com/maranellored/tftpd-go

$ go build

$ ./tftpd-go
```

The server by default runs on port 7125 on the local loopback interface (127.0.0.1)
You can configure the maximum number of threads(go routines) that can be spawned by the server. Each thread is used to handle a single TFTP request. 

You can also manually set the timeout for requests that are made to the server. This value is entered in seconds. 

The default number of go routines is 16 and the default timeout is 1 second.

```
$ ./tftpd-go --port 8080 --threads 32 --timeout 10

$ ./tftpd-go -h
Usage of ./tftpd-go:
  -address="127.0.0.1": The address on which to start the TFTP server
  -port=7125: The port to run the TFTP server on
  -threads=16: The max number of threads to service requests with
  -timeout=1: The timeout for the network connections in seconds
```

The default tftp client that is bundled with Mac OSX can be used to test the server. For example, to use the client 

```
$ echo "Testing TFTP Server" > test

$ echo "Testing TFTP Server, again" > test2

$ echo "put test" | tftp -e 127.0.0.1 7125
Sent 20 bytes in 0.0 seconds

$ echo "get test2" | tftp -e 127.0.0.1 7125
Received 27 bytes in 0.0 seconds
```

**NOTES**: 
* This version doesn't implement retries. 
* This version does not currently perform explicit error checking from the client. If an unexpected packet is encountered from a valid client, it terminates the connection. 
* This version does handle the case where a stray TFTP packet doesn't disrupt an existing transfer

