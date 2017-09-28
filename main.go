package main

import (
	"net"
	"fmt"
	"bufio"
	"flag"
	"os"
	"github.com/sokil/go-connection-pool"
	"log"
	"io/ioutil"
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
	verbose := flag.Bool("verbose", false, "Verbose")
	flag.Parse()

	// configure verbosity of logging
	if *verbose == true {
		log.SetOutput(os.Stderr)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	// create listening socket
	bindAddress := fmt.Sprintf("%s:%d", *host, *port)
	socket, err := net.Listen("tcp", bindAddress)
	if err != nil {
		log.Fatalln("Error starting connectionPool. ", err)
		os.Exit(1)
	} else {
		log.Printf("Server started on %s", bindAddress)
	}

	// channel
	messageChannel := make(chan message, MESSAGE_CHANNEL_LENGTH)

	// prepare connection pool
	connectionPool := connectionPool.NewConnectionPool()

	// accept connections
	go acceptConnections(messageChannel, socket, connectionPool)

	// publish request
	publishMessage(messageChannel, connectionPool)
}

// wait new connection on listening socket
func acceptConnections(
	messageChannel chan message,
	socket net.Listener,
	connectionPool connectionPool.ConnectionPool,
) {
	// wait to accept connection
	for {
		// accept connection
		connection, err := socket.Accept()
		if err != nil {
			log.Fatalln("Error accepting connection. ", err)
			break
		}

		// add connection to pool
		connectionId := connectionPool.Add(connection)

		// log about connection status
		log.Printf("Connection accepted. Connections in pool: %d", connectionPool.Size())

		// handle request
		go waitMessage(
			messageChannel,
			connectionPool,
			connectionId,
		)
	}
}

// wait message in accepted connection
func waitMessage(
	messageChannel chan message,
	connectionPool connectionPool.ConnectionPool,
	connectionId int,
) {
	// create reader
	reader := bufio.NewReader(connectionPool.Get(connectionId))

	// wait for next read buffer ready
	for {
		// read request
		request, _, err := reader.ReadLine()
		if err != nil {
			// stop reading buffer and exit goroutine
			connectionPool.Remove(connectionId)
			log.Printf("Can't read line from socket: %s. Connections in pool: %d", err, connectionPool.Size())
			break
		} else {
			// check request before send
			if len(request) == 0 {
				continue;
			}
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
	connectionPool connectionPool.ConnectionPool,
) {
	// get request from channel
	for {
		// get request from channel
		message := <-messageChannel

		// log accepted message
		log.Printf("Accepted message '%s' on socket %d", message.payload, message.sourceConnectionId)

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
