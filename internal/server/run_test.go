package server

import (
	"io"
	"net"
	"testing"
	"time"

	"tcp-load-balancer/internal/upstream"
)

// maxWaitTime is used to control how long we wait for a connection count change to occur.
const maxWaitTime = time.Second * 5

func TestLoadBalancer_handleConnection_Counter(t *testing.T) {
	t.Run("connection count is incremented and decremented during connection", func(t *testing.T) {

		host, err := upstream.New(":0", "tcp", 1)
		if err != nil {
			t.Error(err)
		}

		l := &LoadBalancer{
			hosts: []*upstream.TcpHost{host},
		}

		// Set up channels to watch for the connection counts to change.
		incremented := make(chan bool)
		decremented := make(chan bool)

		// Start watching the connection count, and allow some time to observe the expected change.
		observeConnectionChange(host, maxWaitTime, 1, incremented)

		if err := l.handleConnection(&net.TCPConn{}); err != nil {
			t.Errorf("LoadBalancer.handleConnection() error = %v", err)
		}

		if !<-incremented {
			t.Error("The host never had its connection count incremented.")
		} else {
			// else the incremented channel returned true, so the connection count has been incremented at this point, it should be decremented soon.
			observeConnectionChange(host, maxWaitTime, -1, decremented)
			if !<-decremented {
				t.Error("The host never had its connection count decremented.")
			}
		}
	})
}

// observeConnectionChange is a helper function that watches the connection count of a host and returns true when expected difference is observed.
// results are written to the input channel.
func observeConnectionChange(host *upstream.TcpHost, timeout time.Duration, expectedDifference int, c chan bool) {
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

// func TestLoadBalancer_handleConnection_Health(t *testing.T) {

// 	healthyHost := createHostWithNConnections(t, 2)
// 	unhealthyHost := &upstream.TcpHost{}

// 	type fields struct {
// 		hosts []*upstream.TcpHost
// 	}
// 	type args struct {
// 		clientConn net.Conn
// 	}
// 	tests := []struct {
// 		name                   string
// 		fields                 fields
// 		args                   args
// 		wantErr                bool
// 		expectedUnhealthyHosts int
// 	}{
// 		{
// 			name: "unhealthy hosts are marked as unhealthy, and the connection is forwarded to the next healthy host",
// 			fields: fields{
// 				// The first host is healthy, and has two active connections, the second is secretly unhealthy and has 0 connections, so it will be selected first.
// 				hosts: []*upstream.TcpHost{healthyHost, unhealthyHost},
// 			},
// 			args: args{
// 				clientConn: &net.TCPConn{},
// 			},
// 			expectedUnhealthyHosts: 1,
// 			wantErr:                false,
// 		},
// 	}
// 	t.Run(tt.name, func(t *testing.T) {
// 		l := &LoadBalancer{
// 				hosts: tt.fields.hosts,
// 			}
// 			if err := l.handleConnection(tt.args.clientConn); (err != nil) != tt.wantErr {
// 				t.Errorf("LoadBalancer.handleConnection() error = %v, wantErr %v", err, tt.wantErr)
// 			}

// 			newUnhealthyHost := make(chan bool)
// 			observeUnhealthyHostChange(l, maxWaitTime, tt.expectedUnhealthyHosts, newUnhealthyHost)

// 			if !<-newUnhealthyHost {
// 				t.Errorf("LoadBalancer.handleConnection() expected %d unhealthy hosts, got %d", tt.expectedUnhealthyHosts, len(l.unhealthyHosts))
// 				hosts: tt.hosts,
// 			}
// 			if err := l.handleConnection(tt.clientConn); (err != nil) != tt.wantErr {
// 				t.Errorf("LoadBalancer.handleConnection() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if !connectionIsClosed(tt.clientConn) {
// 				t.Error("The connection was not properly closed.")
// 			}
// 		})
// 	}
// }
// func TestLoadBalancer_handleConnection_Resources(t *testing.T) {

// 	tests := []struct {
// 		name       string
// 		clientConn net.Conn
// 		hosts      []*upstream.TcpHost
// 		wantErr    bool
// 	}{
// 		{
// 			name:    "connection is closed when LeastConnections returns an error",
// 			wantErr: true,
// 		},
// 		{
// 			name:    "connection is closed after forwarding to host",
// 			hosts:   []*upstream.TcpHost{{}},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			l := &LoadBalancer{
// 				hosts: tt.hosts,
// 			}
// 			if err := l.handleConnection(tt.clientConn); (err != nil) != tt.wantErr {
// 				t.Errorf("LoadBalancer.handleConnection() error = %v, wantErr %v", err, tt.wantErr)
// 			}
// 			if !connectionIsClosed(tt.clientConn) {
// 				t.Error("The connection was not properly closed.")

// 			}
// 		})
// 	}
// }

// observeUnhealthyHostChange is a helper function that watches the unhealthy hosts and returns true when lengthDelta is met.
// results are written to the input channel.
func observeUnhealthyHostChange(l *LoadBalancer, timeout time.Duration, lengthDelta int, c chan bool) {
	startingTime := time.Now()
	startingCount := l.unhealthyHosts.Length()

	go func() {
		for {
			if l.unhealthyHosts.Length() == startingCount+lengthDelta {
				c <- true
			}
			if time.Since(startingTime) > timeout {
				c <- false
			}
		}
	}()
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
