package main

import (
	"net"
	"fmt"
	"bufio"
	"flag"
	"os"
	"./server"
)

const DEFAULT_HOST = "127.0.0.1"
const DEFAULT_PORT = 42042
const MESSAGE_CHANNEL_LENGTH = 10

type message struct {
	payload string
	sourceConnectionId int
}

func main() {
	// declare command line options
	host := flag.String("host", DEFAULT_HOST, "Host")
	port := flag.Int("port", DEFAULT_PORT, "Port")
	flag.Parse()

	// create listening socket
	bindAddress := fmt.Sprintf("%s:%d", *host, *port)
	socket, err := net.Listen("tcp", bindAddress)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	} else {
		fmt.Println(fmt.Sprintf("Server started on %s:%d", *host, *port))
	}

	// channel
	channel := make(chan message, MESSAGE_CHANNEL_LENGTH)

	// prepare connection pool
	connectionPool := server.NewConnectionPool()

	// accept connections
	go acceptConnections(channel, socket, connectionPool)

	// publish request
	publishMessage(channel, connectionPool)
}

// wait new connection on listening socket
func acceptConnections(
	messageChannel chan message,
	socket net.Listener,
	connectionPool server.ConnectionPool,
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
		connectionId := connectionPool.Add(connection)

		// handle request
		go waitMessage(messageChannel, connection, connectionId)
	}
}

// wait message in accepted connection
func waitMessage(
	messageChannel chan message,
	connection net.Conn,
	connectionId int,
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
			messageChannel <- message {
				payload: string(request),
				sourceConnectionId: connectionId,
			}
		}
	}
}

// send message to all subscribers
func publishMessage(
	messageChannel chan message,
	connectionPool server.ConnectionPool,
) {
	// get request from channel
	for {
		// get request from channel
		message := <-messageChannel

		// send request to all connections
		go func() {
			// send response
			connectionPool.Range(func(targetConnection net.Conn, targetConnectionId int) {
				if message.sourceConnectionId == targetConnectionId {
					return
				}
				writer := bufio.NewWriter(targetConnection)
				writer.WriteString(string(message.payload) + "\n")
				writer.Flush()
			})
		}()
	}
}
