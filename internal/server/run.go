package server

import (
	"errors"
	"io"
	"log"
	"net"
	"tcp-load-balancer/internal/upstream"
)

var ErrUninitialized = errors.New("load balancer not initialized")

// Run handles incoming connections until terminated.
func (l *LoadBalancer) Run() error {
	if l == nil || l.listener == nil {
		return ErrUninitialized
	}

	for {
		// TODO: Set up mTLS in next PR. For now, connect without TLS.
		clientConn, err := l.listener.Accept()
		if err != nil {
			log.Printf("Load balancer unexpectedly declined connection and will be shut down: %s", err)
			// TODO: attempt to re-establish the listener with a retry mechanism (leaving out of scope for this project).
			break
		}

		if err := l.handleConnection(clientConn); err != nil {
			log.Printf("Unable to handle connection: %s", err)
		}
	}

	return nil
}

// handleConnection selects an upstream host, tracks connection counts, and forwards data upstream.
func (l *LoadBalancer) handleConnection(clientConn net.Conn) error {
	// Host selection is not included in goroutine handling so that requests arriving at the same time are not routed to the same host.
	// This adds a small amount of latency to the request, but ensures accurate load balancing.
	host, err := l.LeastConnections()
	if err != nil {
		closeConnection(clientConn)
		return err
	}

	// Increment the connection count for the selected host.
	host.IncrementActiveConnections()

	// Copy data to the selected host, and decrement the connection count when the copy finishes.
	go func() {

		if err = forwardToHost(clientConn, host); err != nil {
			// TODO: clean up this nested stuff
			if err == upstream.ErrUnhealthy {

				// If the host is unhealthy, remove it so that leastConnections doesn't select this host again until it's healthy.
				l.TrackUnhealthyHost(host.ID())

				// Start over to select a new host.
				err = l.handleConnection(clientConn)
				if err != nil {
					log.Printf("error when ra-attempting failed host: %s", err)
				}
			}

			host.DecrementActiveConnections()

			return
		}

		closeConnection(clientConn)

		// TODO: Support response data from hosts back to clients (outside scope of this project).

		// Decrement the connection count for the selected host.
		host.DecrementActiveConnections()
	}()

	return nil
}

// forwardToHost copies data from the client to the host.
func forwardToHost(clientConn net.Conn, host *upstream.TcpHost) error {
	hostConn, err := host.Dial()
	if err != nil {
		return err
	}
	defer hostConn.Close()

	// Copy data to host (dst) from client (src). This will stay open until clientConn is closed.
	// Currently allowing the client to decide when to terminate the connection.
	if _, err = io.Copy(hostConn, clientConn); err != nil {
		return err
	}

	return nil
}

// closeConnection closes the connection and logs the error, if any.
func closeConnection(conn net.Conn) {
	if conn == nil {
		return
	}
	if err := conn.Close(); err != nil {
		log.Printf("Unable to close client connection: %s", err)
	}
}
