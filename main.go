package main

import (
	"net"
	"fmt"
	"bufio"
	"flag"
	"os"
	"./pubsub"
)

const DEFAULT_HOST = "127.0.0.1"
const DEFAULT_PORT = 42042

// wait new connection on listening socket
func acceptConnections(
	channel chan string,
	socket net.Listener,
	connectionPool pubsub.ConnectionPool,
) {
	// wait to accept connection
	for {
		// accept connection
		connection, err := socket.Accept()
		if err != nil {
			fmt.Print("Error accepting connection", err)
			break
		}

		// add connection to pool
		connectionPool.Add(connection)

		// handle request
		go waitMessage(channel, connection)
	}
}

// wait message in accepted connection
func waitMessage(
	channel chan string,
	connection net.Conn,
) {
	// create reader
	reader := bufio.NewReader(connection)

	// wait for next read buffer ready
	for {
		// check if connection still alive
		// ...

		// read request
		request, _, err := reader.ReadLine()
		if err != nil {
			// stop reading buffer and exit goroutine
			break
		} else {
			// send request to channel
			channel <- string(request)
		}
	}
}

// send message to all subscribers
func publishMessage(
	channel chan string,
	connectionPool pubsub.ConnectionPool,
) {
	// get request from channel
	for {
		// get request from channel
		request := <-channel

		// send request to all connections
		go func() {
			// send response
			connectionPool.Iterate(func(targetConnection net.Conn, targetConnectionId int) {
				// todo: make resizeable connection pool
				if targetConnection == nil {
					return
				}
				//if connectionId == targetConnectionId {
				//	continue
				//}
				fmt.Println(fmt.Sprintf("Sending to connection %d", targetConnectionId))
				writer := bufio.NewWriter(targetConnection)
				writer.WriteString(string(request) + "\n")
				writer.Flush()
			})
		}()
	}
}

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
	channel := make(chan string, 10)

	// prepare connection pool
	connectionPool := pubsub.NewConnectionPool()

	// accept connections
	go acceptConnections(channel, socket, connectionPool)

	// publish request
	publishMessage(channel, connectionPool)
}
