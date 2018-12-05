# Virtual Chain Gossip Discovery

## Problem

Virtual chains need to be able to connect to their peers, but they can't have fixed ports or IPs assigned to them because we would have to coordinate them between all federation members through some external configuration.

The problem would be more pronounced if we abandon current Docker Swarm networking mode in favor of direct TCP connections. If the Swarm worker containing the node goes down, who would update this external configuration? How would it propagate through the network?

## Solution

Create a discovery service that allows a vchain to discover its peers by querying another node. We already have the list of federation members public keys and their IPs.

## Implementation

`gossip-discovery` service should be hosted by the node and provisioned by Boyar.

`http-api-reverse-proxy` service should be used to proxy calls to `gossip-discovery` service.

HTTP query `/discovery/{vcid}` should return IP and port that will be used by the node to connect to the peer vchain.

> Let's say we have vchain `42` on nodes `A`, `B`, and `C`. After starting up, vchain `A.42` would query `A.gossip-discovery` via an HTTP call, get IP and port combination for `B.42` and will be able to connect to its peer.