package main

import (
	"net"
	"fmt"
	"strconv"
	"bufio"
)

func main() {
	port := 8080

	// create listening socket
	socket, err := net.Listen("tcp", ":" + strconv.Itoa(port))
	if err != nil {
		fmt.Print("Error starting server")
	}

	// accept connections
	for {
		fmt.Println("Wait for incoming connection")
		connection, err := socket.Accept()
		if err != nil {
			fmt.Print("Error accepting connection")
		}
		fmt.Println("Connection accepted")

		// read request
		request, _, err:= bufio.NewReader(connection).ReadLine()

		// send response
		writer:= bufio.NewWriter(connection)
		writer.WriteString("Request: " + string(request) + "\n")
		writer.Flush()

		// close connection
		connection.Close()
	}
}
