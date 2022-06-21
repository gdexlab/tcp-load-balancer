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
			for {
				// TODO: In future PR, use DialTLS to connect securely to the LB.
				conn, err := net.Dial(config.TCPNetwork, address)
				if err != nil {
					log.Print(err)
				}

				log.Printf("Static client at %s sending '%s' to LB address: %s", conn.LocalAddr(), helloMessage, conn.RemoteAddr())

				if _, err := conn.Write([]byte(helloMessage + "\n")); err != nil {
					log.Print(err)
				}

				if err = conn.Close(); err != nil {
					log.Print(err)
				}

				time.Sleep(clientMessageInterval)
			}
		}()
	}
}
