package config

import (
	"flag"
	"strconv"
	"time"
)

const (
	// selectOpenPort will pick an available random port for TCP connections.
	SelectOpenPort = ":0"
	// tcpNetwork could eventually be one of "tcp", "tcp4", "tcp6", but this project currently only supports "tcp".
	TCPNetwork = "tcp"
	// HostFailureThreshold controls the number of failures allowed before a host is marked unhealthy.
	HostFailureThreshold = 3

	// TODO: The following constants are purely for controlling static host/client setup and could be removed in a future version of this app.
	// ------ Start Static Host/Client Config ------
	// numberOfHosts controls the number of static hosts added to the load balancer during setup.
	NumberOfHosts = 5
	// numberOfHosts controls the number of static clients added to the load balancer during setup.
	NumberOfClients = 7
	// clientMessageInterval controls how often messages are sent from the client to the load balancer.
	ClientMessageInterval = time.Second * 3
	// ------ End Static Host/Client Config ------
)

// GetPort returns the port number to listen on. If no port flag is set, it returns default for finding an available port, which is ":0"
func GetPort() string {
	port := flag.Int("p", 0, "Port for the load balancer to listen on")
	flag.Parse()
	return ":" + strconv.Itoa(*port)
}
