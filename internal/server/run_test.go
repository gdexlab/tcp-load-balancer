package server_test

// Using separate _test package to avoid circular dependency with import of "tcp-load-balancer/test" package.

import (
	"net"
	"strings"
	"testing"
	"time"

	"tcp-load-balancer/internal/server"
	"tcp-load-balancer/internal/upstream"
	"tcp-load-balancer/test"

	"github.com/google/uuid"
)

func Test_ForwardData(t *testing.T) {
	uniquePayload := uuid.New().String()
	tests := []struct {
		name                 string
		hostConn             net.Conn
		clientConnectionPipe func() (net.Conn, net.Conn)
		payload              string
		wantErr              bool
	}{
		{
			name: "Data is forwarded to host, and response includes original payload",
			hostConn: func() net.Conn {
				h, err := test.InitializeHost("tcp", ":0")
				if err != nil {
					t.Fatal(err)
				}

				// Connect LB to Host.
				hostConn, err := net.Dial("tcp", h.Addr().String())
				if err != nil {
					t.Fatal(err)
				}

				return hostConn
			}(),
			clientConnectionPipe: net.Pipe,
			payload:              uniquePayload,
			wantErr:              false,
		},
		{
			name: "disconnected client connection results in error",
			clientConnectionPipe: func() (net.Conn, net.Conn) {
				cc, cs := net.Pipe()
				cc.Close()
				return cc, cs
			},
			wantErr: true,
		},
		{
			name: "host closed early results in immediate error",
			clientConnectionPipe: func() (net.Conn, net.Conn) {
				cc, cs := net.Pipe()
				return cc, cs
			},
			hostConn: func() net.Conn {
				hostConn, remoteHostConn := net.Pipe()
				remoteHostConn.Close()

				return hostConn
			}(),
			payload: uniquePayload,
			wantErr: true,
		},
		{
			name:                 "Nil host results in error",
			clientConnectionPipe: net.Pipe,
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Need control over both the client side and server side of the client connection.
			clientConnToLB, lbConnToClient := tt.clientConnectionPipe()

			response := make(chan string, 1)
			go func() {
				if !connectionIsClosed(clientConnToLB) {
					response <- writeAndReadResponse(t, clientConnToLB, tt.payload)
				}
			}()

			if err := server.ForwardData(lbConnToClient, tt.hostConn, time.Second*1); (err != nil) != tt.wantErr {
				t.Errorf("forwardData() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr {
				// Don't wait for response if we are expecting an error.
				return
			}

			result := <-response
			if !strings.Contains(result, tt.payload) {
				t.Errorf("Actual response did not contain original payload: %s does not contain %s", result, tt.payload)
			}
		})
	}
}

// connectionIsClosed is a helper function that checks if a connection has been closed.
func connectionIsClosed(conn net.Conn) bool {
	if conn == nil {
		return true
	}

	_, err := conn.Write([]byte("should fail"))

	// Would like to rely on exact error here, but the error is not exported, so if Write returns an error, we'll treat that as a closed connection.
	// Broken pipe (not exported) or io.EOF are the error cases we usually see when connection is closed.
	// `write tcp [::1]:4333->[::1]:50403: write: broken pipe`
	return err != nil
}

func TestLoadBalancer_handleConnection_Counter(t *testing.T) {
	t.Run("connection count is incremented and decremented during connection", func(t *testing.T) {

		h, err := test.InitializeHost("tcp", ":0")
		if err != nil {
			t.Fatal(err)
		}

		host, err := upstream.New(h.Addr().String(), "tcp")
		if err != nil {
			t.Fatal(err)
		}

		l, err := server.New("tcp", ":0", time.Second*1)
		if err != nil {
			t.Fatal(err)
		}

		l.AddUpstream(host)

		clientConn, clientServerConn := net.Pipe()

		// Set up channels to watch for the connection counts to change.
		incremented := make(chan bool)
		decremented := make(chan bool)

		// Start watching the connection count, and allow some time to observe the expected change.
		expectConnectionChange(host, time.Second*5, 1, incremented)

		if err := l.HandleConnection(clientServerConn); err != nil {
			t.Errorf("LoadBalancer.HandleConnection() error = %v", err)
		}

		_ = writeAndReadResponse(t, clientConn, "test")
		if !<-incremented {
			t.Error("The host never had its connection count incremented.")
		} else {
			// else the incremented channel returned true, so the connection count has been incremented at this point, it should be decremented soon.
			expectConnectionChange(host, time.Second*5, -1, decremented)
			if !<-decremented {
				t.Error("The host never had its connection count decremented.")
			}
		}

		if !connectionIsClosed(clientServerConn) {
			t.Error("The connection was not properly closed.")
		}
	})
}

// expectConnectionChange is a helper function that watches the connection count of a host and returns true when expected difference is observed.
// results are written to the input channel.
func expectConnectionChange(host *upstream.TcpHost, timeout time.Duration, expectedDifference int, c chan bool) {
	startingTime := time.Now()
	startingCount := int(host.ConnectionCount())
	go func() {
		for {
			if int(host.ConnectionCount()) == startingCount+expectedDifference {
				c <- true
				return
			}
			if time.Since(startingTime) > timeout {
				c <- false
				return
			}
		}
	}()
}

// writeAndReadResponse is a helper to simulate client activities of writing to a connection and returning the response.
func writeAndReadResponse(t *testing.T, conn net.Conn, payload string) string {

	// Write data from client.
	_, err := conn.Write([]byte(payload))
	if err != nil {
		t.Error(err)
	}

	// Receive response on client.
	response := make([]byte, 1024)
	n, err := conn.Read(response)
	if err != nil {
		t.Error(err)
	}
	conn.Close()

	return string(response[:n])
}
