package server

import (
	"errors"
	"io"
	"log"
	"net"
	"time"
)

var ErrUninitialized = errors.New("load balancer not initialized")
var ConnectionNotEstablished = errors.New("net.Conn cannot be nil")

// Run handles incoming connections until terminated.
func (l *LoadBalancer) Run() error {
	if l.listener == nil {
		return ErrUninitialized
	}

	for {
		// TODO: Set up mTLS in next PR. For now, connect without TLS.
		clientConn, err := l.listener.Accept()
		if err != nil {
			// TODO: attempt to re-establish the listener with a retry mechanism (leaving out of scope for this project).
			return err
		}

		if err := l.HandleConnection(clientConn); err != nil {
			log.Printf("Unable to handle connection: %s", err)
		}
	}
}

// handleConnection selects an upstream host, tracks connection counts, and forwards data upstream.
func (l *LoadBalancer) HandleConnection(clientConn net.Conn) error {
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

		hostConn, err := host.Dial()
		if err != nil {
			// TODO: Select a different host if this host is down (next PR).
			log.Printf("Error dialing host: %s", err)
			closeConnection(clientConn)
			return
		}

		if err = ForwardData(clientConn, hostConn, l.hostTimeout); err != nil {
			// TODO: Select a different host if this host is down, and communicate the error over a channel rather than just logging it here (next PR).
			log.Printf("Error forwarding data: %s", err)
		}

		closeConnection(clientConn)
		closeConnection(hostConn)

		// Decrement the connection count for the selected host.
		host.DecrementActiveConnections()
	}()

	return nil
}

// ForwardData copies data from the client to the host, and also from the host to the client.
// It does not enforce a timeout on the client, but will timeout if the host is non-responsive.
func ForwardData(clientConn net.Conn, hostConn net.Conn, hostTimeout time.Duration) error {
	if clientConn == nil || hostConn == nil {
		return ConnectionNotEstablished
	}

	hostErr := make(chan error, 1)

	go func() {
		// Copy response from host to client. It will continue running until hostConn is closed.
		hostConn.SetReadDeadline(time.Now().Add(hostTimeout))
		_, err := io.Copy(clientConn, hostConn)
		hostErr <- err
	}()

	// Copy data to host (dst) from client (src). This will stay open until clientConn is closed.
	if _, err := io.Copy(hostConn, clientConn); err != nil {
		return err
	}

	return <-hostErr
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
