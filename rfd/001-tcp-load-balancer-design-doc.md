---
authors: Grant Dexter (gdexlab), Tim Ross (rosstimothy), Marek Smoliński (smallinsky), Tim Buckley (timothyb89)
state: in progress
---
# RFD 1 - Design for TCP Load Balancer

## Overview

A basic TCP load balancer to distribute network load across multiple host servers (referred to as "upstreams"). With requirements pulled from level 5 of the [gravitational career challenge](https://github.com/gravitational/careers/blob/rjones/challenge-2.md/challenges/systems/challenge-2.md).

The load balancer will be written in Go, leveraging primarily standard libraries, and following style principles from [effective go](https://go.dev/doc/effective_go) and [common go code review comments](https://github.com/golang/go/wiki/CodeReviewComments).

The goal is to write as little code as possible, while still satisfying requirements. This will mean that clients and upstream hosts will be statically configured, connection information will be entirely in-memory (not fault-tolerant if LB unexpectedly died), and other areas of scope will be reduced to less than production-ready solutions in the interest of time. Such short-cuts will be noted in comments where applicable. The service should be considered a minimum viable product, while still implementing sufficient testing and verification of functionality.

## Requirements

The following diagram offers an simplified overview of the required steps taken by the load balancer, upstream hosts, and downstream clients. It provides a quick view of the process from when a client initiates connection to when the data is passed to the upstream host. Each "swimlane" at the top of the diagram indicates which entity performs the action. See the following sections beneath the diagram for more details on each step.

<img src="tcp_load_balancer.png" alt="tcp load balancer diagram" width="600"/>


### Authentication

In order to connect to the Load Balancer, mTLS authentication is required. The standard go "crypto/tls" library "partially implements TLS 1.2, as specified in RFC 5246, and TLS 1.3, as specified in RFC 8446." This project will likely leverage that library. Additionally, this design references openssl for generating certificates and keys, which will also support TLS 1.3. For simplicity, this project can focus on TLS 1.3, and then a future iteration could include older versions (depending on what clients need to use this load balancer, and what those clients support). Clients attempting to connect with an older version of TLS will receive an error response. The following actions will be taken to authenticate a client.

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

# Generate clientca.cert and clientcakey.pem to allow signing of client keys.
$ openssl genrsa -out clientcakey.pem
$ openssl req -new -x509 -key clientcakey.pem -out clientca.cert

Generate private (client.crt) and public (client.key) keys for the client:
$ openssl genrsa -out client.key
$ openssl req -new -key client.key -out client_reqout.txt
$ openssl x509 -req -in client_reqout.txt -days 3650 -sha256 -CAcreateserial -CA clientca.cert -CAkey clientcakey.pem -out client.crt		
```

For this project, a test client will be created and provided with a valid client cert for testing purposes. These initial dummy certificates and keys will be included (as plaintext) within this repository for the the benefit of reviewers, but should stored as secure secrets and managed outside of this repo in a future version.

Additionally, the requirement of over-the-wire encryption is not mentioned, so even though the authentication requires certs and keys, the data sent after authentication may or may not be encrypted depending on what is most convenient at the point of implementation.

### Rate Limiter
This load balancer will implement a per-client connection rate limiter that tracks the number of client connections, and limits to n (configurable) active connections per client. It could track client identity based on attributes like IP address, and device identifier, but since this load balancer requires authentication (mTLS), we will identify unique clients based on a V5 UUID generated from the client's TLS certificate. This can be done leveraging the `crypto/tls` library, which makes peer certificates available on the `tls.Conn` struct returned from `tls.Client`. These can be turned into a consistent V5 uuid, to be used for `ClientID`, by creating a hash from the certificate, as shown below.

```
// pseudo code -- requires changes during implementation, but demonstrates proof of concept.

conn, err := l.listener.Accept()
tlsConn := tls.Client(conn, tls.Config)
cert := tlsConn.ConnectionState().PeerCertificates[0]

clientID := uuid.NewV5(uuid.NamespaceOID, cert)
```

Client connection counts can be stored in a `sync.Map` structure for quick modification, and safety across concurrent goroutines. Regarding this map choice, the [`sync` package documentation](https://pkg.go.dev/sync#Map) states the following: 
> "The Map type is optimized for two common use cases: (1) when the entry for a given key is only ever written once but read many times, as in caches that only grow, or (2) when multiple goroutines read, write, and overwrite entries for disjoint sets of keys. In these two cases, use of a Map may significantly reduce lock contention compared to a Go map paired with a separate Mutex or RWMutex."

Our use-case follows the #2 reason for leveraging `sync.Map` over a standard map with a Mutex; our disjoint sets of keys will be the unique client ids.



### Upstream Authorization Management

The load balancer will employ a simple authorization scheme which will deny access to specific upstreams for certain clients. This scheme will be statically defined in code, using something like the following `map` of `UnauthorizedClientHostRule`s to ensure an authorized host is selected. To simply demonstrate this capability, we store a map of `clientID:hostID` as an `UnauthorizedClientHostRule`. The only example rule which will be present for this project will be that client1 can never talk to host2. 

```

// UnauthorizedClientHostRule is a composite key made up of ClientID and HostID, and represents that the client is not authorized to connect to the host.
type UnauthorizedClientHostRule string

// FormatUnauthorizedClientHostRule builds a new unauthorizedClientHostRule from the given client and host IDs.
func BuildUnauthorizedClientHostRule(clientID uuid.UUID, hostID uuid.UUID) UnauthorizedClientHostRule {
	return UnauthorizedClientHostRule(fmt.Sprintf("%s:%s", clientID, hostID))
}

// unauthorized serves as a key-based lookup for unauthorized client-> host relationships.
// It will be checked every time a host is being selected. 
// If upstream hosts are to be registered dynamically, access to this map should be locked behind a mutex.
// For simplicity of this project, we'll only ever add rules during app startup, so the map is safe for concurrent reads.
var unauthorized = map[UnauthorizedClientHostRule]struct{}{}
```

Once the load balancer has determined that the client is authenticated and within client connection limit (and prior to selecting the least connections host) it will filter out unauthorized hosts for that client. See the following section for how the unauthorized map may be used for connection filtering.

### Request Forwarder

Incoming packets will be forwarded after checking 3 criterion:
* Client must be authenticated.
* Client must be within connection limit.
* Client must be authorized to access at least one healthy host.

Once these requirements are satisfied, the request will be forwarded to an authorized host with the least connections.

The least connection tracking will prioritize simplicity over performance for this project. Each new connection that is made will increment a host's connection count, and each completed connection will decrement the host's connection count. Each incoming request will iterate through all authorized hosts to find the one with the least active connections. A more performant solution could later be devised which may not have to iterate over every authorized host for every request (likely leveraging an async queue). In order to handle the edge case where a large batch of requests arrive at exact same time, a mutex will be leveraged to lock access to the `LeastConnections` and `IncreaseConnectionCount` methods; the increased latency from this locking step should be minimal.

Example of simple approach for selecting host:
```
func (l *LoadBalancer) LeastConnections(clientID uuid.UUID) *upstream.TcpHost {
	if l == nil || len(l.hosts) == 0 {
		return nil
	}

	var host *upstream.TcpHost

	for i, h := range l.hosts {
     _, unauthzd := unauthorized[BuildUnauthorizedClientHostRule(clientID, h.ID)]
		if !unauthzd && (host == nil || h.ConnectionCount() < host.ConnectionCount()) {
			host = h
		}
	}

	return host
}
```

Once the host is selected, data will be passed directly from clients to upstream hosts, without being stored on the load balancer.

### Health Checks
The load balancer removes unhealthy upstreams if it is unable to connect to the host while handling a client connection. The health checks could include a grace period, or some request-forwarding backoff, allowing for n attempts before removal, but for simplicity, this project will immediately remove the host from available hosts upon a single failed connection.

The load balancer will store a registry of all hosts and with health status for each host. A separate go routing will periodically recheck unhealthy hosts so that statuses can be updated when health is restored. A host which allows connections will be considered healthy.

## Demonstrating Functionality

The current plan (subject to change with implementation) to demonstrate successful functionality of the project will be to initialize and connect a few hosts and a client as part of the main.go file for the load balancer. This should provide a quick way to test local development while iterating on a solution. Log statements will output actions for simple observability. The app will be run with `go run main.go`. In a future version, the service would either allow configuration for upstream hosts, or expose methods for registering new upstream hosts, but that will be out of scope for now.

A random available port will be selected to start up the service unless a port is passed with `-p` as a flag. Manual testing/access could be done via tcp utilities like telnet if desired, e.g. `telnet localhost [PORTNUMBER]`

Additionally, basic unit tests will be included to validate key features.