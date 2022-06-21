package server

import (
	"io"
	"net"
	"tcp-load-balancer/internal/upstream"
	"testing"
	"time"
)

func TestLoadBalancer_handleConnection_Counter(t *testing.T) {
	t.Run("connection count is incremented and decremented during connection", func(t *testing.T) {

		host := &upstream.TcpHost{}
		l := &LoadBalancer{
			hosts: []*upstream.TcpHost{host},
		}

		// Set up channels to watch for the connection counts to change.
		incremented := make(chan bool)
		decremented := make(chan bool)

		// Start watching the connection count, and allow some time to observe the expected change.
		expectConnectionChange(host, time.Second*5, 1, incremented)

		if err := l.handleConnection(&net.TCPConn{}); err != nil {
			t.Errorf("LoadBalancer.handleConnection() error = %v", err)
		}

		if !<-incremented {
			t.Error("The host never had its connection count incremented.")
		} else {
			// else the incremented channel returned true, so the connection count has been incremented at this point, it should be decremented soon.
			expectConnectionChange(host, time.Second*5, -1, decremented)
			if !<-decremented {
				t.Error("The host never had its connection count decremented.")
			}
		}
	})
}

// expectConnectionChange is a helper function that watches the connection count of a host and returns true when expected difference is observed.
// results are written to the input channel.
func expectConnectionChange(host *upstream.TcpHost, timeout time.Duration, expectedDifference int, c chan bool) {
	startingTime := time.Now()
	startingCount := host.ConnectionCount()
	go func() {
		for {
			if host.ConnectionCount() == startingCount+expectedDifference {
				c <- true
			}
			if time.Since(startingTime) > timeout {
				c <- false
			}
		}
	}()
}

func TestLoadBalancer_handleConnection_Resources(t *testing.T) {

	tests := []struct {
		name       string
		clientConn net.Conn
		hosts      []*upstream.TcpHost
		wantErr    bool
	}{
		{
			name:    "connection is closed when LeastConnections returns an error",
			wantErr: true,
		},
		{
			name:    "connection is closed after forwarding to host",
			hosts:   []*upstream.TcpHost{{}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &LoadBalancer{
				hosts: tt.hosts,
			}
			if err := l.handleConnection(tt.clientConn); (err != nil) != tt.wantErr {
				t.Errorf("LoadBalancer.handleConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !connectionIsClosed(tt.clientConn) {
				t.Error("The connection was not properly closed.")
			}
		})
	}
}

// connectionIsClosed is a helper function that checks if a connection has been properly closed.
func connectionIsClosed(conn net.Conn) bool {
	if conn == nil {
		return true
	}

	conn.SetReadDeadline(time.Now())
	data := make([]byte, 1)
	_, err := conn.Read(data)

	return err == io.EOF
}
