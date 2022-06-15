---
authors: Grant Dexter (gdexlab), Tim Ross (rosstimothy), Marek Smoliński (smallinsky), Tim Buckley (timothyb89)
state: in progress
---
# RFD 1 - Design for TCP Load Balancer

## Overview

A basic TCP load balancer to distribute network load across multiple host servers (referred to as "upstreams"). With requirements pulled from level 5 of the [gravitational career challenge](https://github.com/gravitational/careers/blob/rjones/challenge-2.md/challenges/systems/challenge-2.md).

The load balancer will be written in Go, leveraging primarily standard libraries, and following style principles from [effective go](https://go.dev/doc/effective_go) and [common go code review comments](https://github.com/golang/go/wiki/CodeReviewComments).

The goal is to write as little code as possible, while still satisfying requirements. The service should be considered a minimum viable product, while still implementing sufficient testing and verification of functionality.

## Requirements

The following diagram offers an simplified overview of the required steps taken by the load balancer, upstream hosts, and downstream clients. It provides a quick view of the process from when a client initiates connection to when the data is passed to the upstream host. Each "swimlane" at the top of the diagram indicates which entity performs the action. See the following sections beneath the diagram for more details on each step.

<img src="tcp_load_balancer.png" alt="tcp load balancer diagram" width="600"/>


### Authentication

In order to connect to the Load Balancer, mTLS authentication is required. The following actions will be taken to authenticate a client.

* Client connects to load balancer (LB)
* LB provides TLS certificate
* Client verifies LB certificate
* Client presents TLS certificate
* LB verifies client certificate

The Load Balancer will have a public and private key as well as a certificate, and the client will also have a public and private key, as well as a certificate. For this project, the certificates and keys will be generated with `openssl` and certificates will be self-signed for simplicity. 

Example steps for certificate and key generation:

```
# Generate serverca.cert and servercakey.pem to allow signing of server keys.
$ openssl genrsa -out servercakey.pem
$ openssl req -new -x509 -key servercakey.pem -out serverca.cert

# Generate public (server.key) and private (server.crt) keys for the the server.
$ openssl genrsa -out server.key
$ openssl req -new -key server.key -out server_reqout.txt
$ openssl x509 -req -in server_reqout.txt -days 3650 -sha256 -CAcreateserial -CA serverca.cert -CAkey servercakey.pem -out server.crt

# Generate clentca.cert and clientcakey.pem to allow signing of client keys.
$ openssl genrsa -out clientcakey.pem
$ openssl req -new -x509 -key clientcakey.pem -out clentca.cert

Generate private (client.crt) and public (client.key) keys for the client:
$ openssl genrsa -out client.key
$ openssl req -new -key client.key -out client_reqout.txt
$ openssl x509 -req -in client_reqout.txt -days 3650 -sha256 -CAcreateserial -CA clentca.cert -CAkey clientcakey.pem -out client.crt		
```

For this project, a test client will be created and provided with a valid client cert for testing purposes. These initial dummy certificates and keys will be included (as plaintext) within this repository for the the benefit of reviewers, but should stored as secure secrets and managed outside of this repo in a future version.

Additionally, the requirement of over-the-wire encryption is not mentioned, so even though the authentication requires certs and keys, the data sent after authentication may or may not be encrypted depending on what is most convenient at the point of implementation.

### Rate Limiter
This load balancer will implement a per-client connection rate limiter that tracks the number of client connections, and limits to n (configurable) active connections per client. It could track client identity based on attributes like IP address, and device identifier, but since this load balancer requires authentication (mTLS), we will identify unique clients based on a V5 UUID generated from the client's TLS certificate. 

```
clientID := uuid.NewV5(loadBalancerID, client.TLSCertificate().String())
```

Client connection counts can be stored in a `sync.Map` structure for quick modification, and safety across concurrent goroutines. Regarding this map choice, the [`sync` package documentation](https://pkg.go.dev/sync#Map) states the following: 
> "The Map type is optimized for two common use cases: (1) when the entry for a given key is only ever written once but read many times, as in caches that only grow, or (2) when multiple goroutines read, write, and overwrite entries for disjoint sets of keys. In these two cases, use of a Map may significantly reduce lock contention compared to a Go map paired with a separate Mutex or RWMutex."

Our use-case follows the #2 reason for leveraging `sync.Map` over a standard map with a Mutex; our disjoint sets of keys will be the unique client ids.


### Upstream Authorization Management

The load balancer will employ a simple authorization scheme which will deny access to specific upstreams for certain clients. This scheme will be statically defined in code. Once the load balancer has determined that the client is authenticated and within client connection limit (and prior to selecting the least connections host) it will filter out unauthorized hosts for that client. To simply demonstrate this capability, we store a map of `clientID`s (based on client certificate) to disallowed `hostID`s; because host configuration is outside the scope of this project, the order in which the host was registered with the load balancer will be used as the `hostID`.

### Request Forwarder

Incoming packets will be forwarded after checking 3 criterion:
* Client must be authenticated.
* Client must be within connection limit.
* Client must be authorized to access at least one healthy host.

Once these requirements are satisfied, the request will be forwarded to an authorized host with the least connections.

The least connection tracking will prioritize simplicity over performance for this project. Each new connection that is made will increment a host's connection count, and each completed connection will decrement the host's connection count. Each incoming request will iterate through all authorized hosts to find the one with the least active connections. A more performant solution could later be devised which may not have to iterate over every authorized host for every request (likely leveraging an async queue). One additional shortcoming of this simple approach is that a large batch of requests which arrive at exact same time could be routed to the exact same host, but this initial project is not planning to handle that edge case.

Example of simple approach for selecting host:
```
func (l *LoadBalancer) LeastConnections() *upstream.TcpHost {
	if l == nil || len(l.hosts) == 0 {
		return nil
	}

	host := l.hosts[0]

	for i, h := range l.hosts[1:len(hosts)] {
		if h.ConnectionCount() < host.ConnectionCount() {
			host = h
		}
	}

	return host
}
```

Once the host is selected, data will be passed directly from clients to upstream hosts, without being stored on the load balancer.

### Health Checks
The load balancer removes unhealthy upstreams if it is unable to connect to the host while handling a client connection. The health checks could include a grace period, or some request-forwarding backoff, allowing for n attempts before removal, but for simplicity, this project will immediately remove the host from available hosts upon a single failed connection.


## Demonstrating Functionality

The current plan (subject to change with implementation) to demonstrate successful functionality of the project will be to initialize and connect a few hosts and a client as part of the main.go file for the load balancer. This should provide a quick way to test local development while iterating on a solution. Log statements will output actions for simple observability. The app will be run with `go run main.go`, and a pre-built binary to work on 64-bit Linux machines will eventually be included once the app is complete. In a future version, the service would either allow configuration for upstream hosts, or expose methods for registering new upstream hosts, but that will be out of scope for now.

A random available port will be selected to start up the service, and manual testing/access could be done via tcp utilities like telnet if desired, e.g. `telnet localhost [PORT DISPLAYED AT STARTUP]`

Additionally, basic unit tests will be included to validate key features.