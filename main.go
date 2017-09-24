package main

import (
	"net"
	"fmt"
	"bufio"
	"flag"
	"os"
)

const DEFAULT_HOST = "127.0.0.1"
const DEFAULT_PORT = 42042

func main() {
	host := *flag.String("host", DEFAULT_HOST, "Host")
	port := *flag.Int("port", DEFAULT_PORT, "Port")

	// create listening socket
	bindAddress := fmt.Sprintf("%s:%d", host, port)
	socket, err := net.Listen("tcp", bindAddress)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	} else {
		fmt.Println(fmt.Sprintf("Server started on %s:%d", host, port))
	}

	// channel
	channel := make(chan string)

	// prepare connection pool
	connectionId := 0
	connectionPool := [10]net.Conn {}

	// accept connections
	fmt.Println("Waiting for incoming connection")
	go func() {
		// wait to accept connection
		for {
			// accept connection
			connection, err := socket.Accept()
			if err != nil {
				fmt.Print("Error accepting connection", err)
				continue
			}

			// add connection to pool
			connectionPool[connectionId] = connection
			connectionId = connectionId + 1

			// handle request
			go func(channel chan string) {
				fmt.Println(fmt.Sprintf("Accepting connection %d", connectionId))

				// create reader
				reader := bufio.NewReader(connection)

				// wait for next read buffer ready
				for {
					// check if connection still alive
					// ...

					// read request
					request, _, err := reader.ReadLine()
					if err != nil {
						fmt.Println("Error reading request stream", err)
						// stop reading buffer and exit goroutine
						break
					} else {
						fmt.Println(fmt.Sprintf("Connection %d accept message: %s", connectionId, string(request)))
						// send request to channel
						channel <- string(request)
					}
				}
			}(channel)
		}
	}()

	// get request from channel
	for {
		// get request from channel
		request := <- channel
		fmt.Println("Request in channel:", request)

		// send request to all connections
		go func() {
			// send response
			for targetConnectionId, targetConnection :=range connectionPool {
				if connectionId == targetConnectionId {
					continue
				}
				fmt.Println(fmt.Sprintf("Sending to connection %d", targetConnectionId))
				writer:= bufio.NewWriter(targetConnection)
				writer.WriteString(string(request) + "\n")
				writer.Flush()
			}
		}()
	}
}
