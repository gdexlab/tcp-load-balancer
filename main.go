package main

import (
	"log"

	"tcp-load-balancer/internal/config"
	"tcp-load-balancer/internal/server"
	"tcp-load-balancer/test"
)

func main() {
	// Initialize the load balancer.
	lb, err := server.New(config.TCPNetwork, config.GetPort())
	if err != nil {
		log.Fatalf("unable to start tcp load balancer: %s", err)
	}

	log.Printf("Load balancer listening on %s", lb.Address())

	// Manually configure upstream hosts and downstream clients to demonstrate functionality.
	if err = test.Setup(lb, config.NumberOfHosts, config.NumberOfClients, config.ClientMessageInterval); err != nil {
		log.Fatalf("unable to setup static connection simulators: %s", err)
	}

	// Await connections until process is terminated.
	if err = lb.Run(); err != nil {
		log.Fatalf("error running tcp load balancer: %s", err)
	}
}
