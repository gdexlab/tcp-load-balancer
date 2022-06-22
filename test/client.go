package test

import (
	"log"
	"net"
	"time"

	"tcp-load-balancer/internal/config"
)

const helloMessage = "Hello"

// InitializeHelloClient is a temporary helper to simulate a client that will connect and pass data to the input address.
// It will create a new connection as frequent as the clientMessageInterval param, and send a "hello" message each time.
func InitializeHelloClients(address string, clientMessageInterval time.Duration, numberOfClients int) {
	for i := 0; i < numberOfClients; i++ {
		go func() {
			ticker := time.NewTicker(clientMessageInterval)
			for range ticker.C {

				// TODO: In future PR, use DialTLS to connect securely to the LB.
				conn, err := net.Dial(config.TCPNetwork, address)
				if err != nil {
					log.Printf("Client was unable to dial: %s", err)
					break
				}

				log.Printf("Static client at %s sending '%s' to LB address: %s", conn.LocalAddr(), helloMessage, conn.RemoteAddr())
				SendThenReceive(conn, helloMessage)

				if err = conn.Close(); err != nil {
					log.Printf("Client was unable to close: %s", err)
				}
			}
		}()
	}
}

// SendThenReceive will send a message to the net.Conn and print off the response it receives.
func SendThenReceive(conn net.Conn, outgoingMessage string) {
	if _, err := conn.Write([]byte(helloMessage)); err != nil {
		log.Printf("Client was unable to write: %s", err)
	}

	// TODO: Outside scope of this project, implement strategy for larger messages.
	data := make([]byte, 2048)
	n, err := conn.Read(data)
	if err != nil {
		log.Printf("client had error reading data: %s", err)
	} else {
		log.Printf("Client at %s received response: %s ", conn.LocalAddr(), string(data[:n]))
	}
}
