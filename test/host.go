package test

import (
	"fmt"
	"io"
	"log"
	"net"
)

// InitializeHost is a temporary helper to simulate an upstream host that respond to acknowledge the data it received.
func InitializeHost(tcpNetwork, address string) (net.Listener, error) {
	h, err := net.Listen(tcpNetwork, address)
	if err != nil {
		return nil, err
	}

	log.Printf("Statically defined host listening on %s", h.Addr())

	go func() {
		// In a goroutine, continuously accept connections.
		for {
			// TODO: In the PR which includes health checks, set up some interval or predictable/controllable behavior where a host will not accept connection so that we can test the health checking process.
			conn, err := h.Accept()
			if err != nil {
				log.Printf("host was unable to accept incoming connection: %s", err)
				return
			}
			go func() {
				for {
					// Continue reading from the established connection until the client closes the connection (resulting in EOF).
					// TODO: Outside scope of this project, implement strategy for larger messages.
					data := make([]byte, 2048)
					n, err := conn.Read(data)
					if err != nil {
						if err != io.EOF {
							log.Printf("error reading data: %s", err)
						}
						return
					}

					// TODO: ensure input data is sanitized.
					_, err = conn.Write([]byte(fmt.Sprintf("Data '%s' was received by host at %s\n", data[:n], h.Addr())))
					if err != nil {
						log.Printf("error writing response: %s", err)
						return
					}
				}
			}()
		}
	}()

	return h, nil
}
