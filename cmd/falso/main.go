package main

import (
	"falso"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
)

var (
	address = flag.String(
		"address", "localhost:8080", "listens for new requests at localhost:8080 by default")

	remoteAddress = flag.String("remoteAddress", "", "define a remote address to proxy")

	dataPath = flag.String("dataPath", "falsoData", "recorded responses will be saved in this path")

	bufferSize = flag.Uint("buffer", 65535, "max buffer size for read")

	mode = flag.String(
		"mode", falso.MOCK, "use proxy/mock, 'mock' will read from the file and send it back and "+
			"'proxy' will try to get a response from the remote address")

	overwrite = flag.Bool("overwrite", false, "use true/false, overwrites existing data in proxy mode")
)

func main() {
	flag.Parse()

	if _, err := os.Stat(*dataPath); os.IsNotExist(err) {
		err = os.Mkdir(*dataPath, 0755)
		if err != nil {
			log.Fatalf("failed to create path %s: %v", *dataPath, err)
		}
	}

	// Listen for incoming connections.
	conn, err := net.Listen("tcp", *address)
	if err != nil {
		log.Fatalf("error listening: %v", err.Error())
	}
	defer func(conn net.Listener) {
		err := conn.Close()
		if err != nil {
			panic(err)
		}
	}(conn)

	mocker := falso.NewMocker(falso.NewDialer(), *mode, *remoteAddress, *dataPath, *bufferSize, *overwrite)
	fmt.Printf("listening on %s in mode %s\n", *address, *mode)
	if *mode == falso.PROXY {
		log.Printf("remote address: %s\n", *remoteAddress)
	}
	for {
		// Listen for an incoming connection.
		conn, err := conn.Accept()
		if err != nil {
			log.Fatalf("error accepting: %v", err.Error())
		}
		// Handle connections in a new goroutine.
		go mocker.HandleRequest(conn)
	}
}
