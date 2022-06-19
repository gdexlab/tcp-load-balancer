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
			go respondAndClose(clientConn, err.Error())
			continue
		}

		if err := l.handleConnection(clientConn); err != nil {
			log.Printf("Unable to handle connection: %s", err)
			go respondAndClose(clientConn, err.Error())
		}
	}
}

// handleConnection selects an upstream host, tracks connection counts, and forwards data upstream.
func (l *LoadBalancer) handleConnection(clientConn net.Conn) error {
	// Host selection is not included in goroutine handling so that requests arriving at the same time are not routed to the same host.
	// This adds a small amount of latency to the request, but ensures accurate load balancing.
	host, err := l.LeastConnections()
	if err != nil {
		return err
	}

	// Increment the connection count for the selected host.
	if err = host.IncrementActiveConnections(); err != nil {
		return err
	}

	// Copy data to the selected host, and decrement the connection count when the copy finishes.
	go func() {
		if err = forwardToHost(clientConn, host); err != nil {
			// TODO: Select a different host if this host is down, and communicate the error over a channel rather than just logging it here (next PR).
			log.Print(err)
		}

		// TODO: Support response data from hosts back to clients (outside scope of this project).

		// Decrement the connection count for the selected host.
		if err = host.DecrementActiveConnections(); err != nil {
			log.Print(err)
			respondAndClose(clientConn, err.Error())
		}
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

// respondAndClose writes a message back to the client and closes the connection.
func respondAndClose(conn net.Conn, message string) {
	if _, err := conn.Write([]byte(message + "\n")); err != nil {
		log.Printf("unable to write to connection: %s", err)
	}

	if err := conn.Close(); err != nil {
		log.Printf("unable to close connection: %s", err)
	}
}
