TFTP
----

A simple tftp server in Go.

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

```
$ ./tftpd-go --port 8080 --threads 32 --timeout 10
```

**NOTE**: This version doesn't implement retries. Does not currently perform explicit error checking from the client.
