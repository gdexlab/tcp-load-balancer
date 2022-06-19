package static

import (
	"time"

	"tcp-load-balancer/internal/config"
	"tcp-load-balancer/internal/server"
	"tcp-load-balancer/internal/upstream"
)

// Setup configures upstream hosts and downstream clients to demonstrate functionality.
func Setup(l *server.LoadBalancer, numberOfHosts int, numberOfClients int, clientMessageInterval time.Duration) error {

	// TODO: Implement a non-static method for registering upstream hosts (outside the scope of this challenge).
	if err := RegisterUpstreamHosts(l, numberOfHosts); err != nil {
		return err
	}

	// TODO: Implement a non-static method for connecting clients (outside the scope of this challenge).
	InitializeHelloClients(l.Address().String(), clientMessageInterval, numberOfClients)

	return nil
}

// RegisterUpstreamHosts adds n static hosts to the load balancer for testing and demonstration purposes.
func RegisterUpstreamHosts(l *server.LoadBalancer, n int) error {
	for i := 0; i < n; i++ {
		h := InitializeHost(l.Address().Network(), config.SelectOpenPort)
		u, err := upstream.New(h.Addr().String(), l.Address().Network())
		if err != nil {
			return err
		}

		l.AddUpstream(u)
	}
	return nil
}
