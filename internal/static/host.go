package static

import (
	"io"
	"log"
	"net"
)

// InitializeHost is a temporary helper to simulate an upstream host that will print off any incoming data.
func InitializeHost(tcpNetwork, address string) net.Listener {
	h, err := net.Listen(tcpNetwork, address)
	if err != nil {
		log.Fatalf("unable to listen on %s: %s", address, err)
	}

	log.Printf("Statically defined host listening on %s", h.Addr())

	go func() {
		// In a goroutine, continuously accept connections.
		for {

			// TODO: In the PR which includes health checks, set up some interval or predictable/controllable behavior where a host will not accept connection so that we can test the health checking process.
			conn, err := h.Accept()
			if err != nil {
				log.Printf("host was unable to accept incoming connection: %s", err)
				// Intentionally allowing the hosts to continue to accept connections for now. This behavior will be refined in the next PR when we set up static hosts to fail occasionally.
				continue
			}

			// Continue reading from the established connection until the client closes the connection (resulting in EOF).
			for {
				// TODO: Outside scope of this project, implement strategy for larger messages.
				data := make([]byte, 2048)
				n, err := conn.Read(data)
				if err != nil {
					if err == io.EOF {
						break
					}
					log.Printf("error reading data: %s", err)
				}

				log.Printf("Host at %s received data: %s", h.Addr(), data[:n])
			}
		}
	}()

	return h
}
