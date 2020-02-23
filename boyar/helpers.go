package boyar

import (
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
)

func buildPeersMap(nodes []*strelets.FederationNode, gossipPort int) *PeersMap {
	peersMap := make(PeersMap)

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		peersMap[config.NodeAddress(node.Address)] = &Peer{
			node.IP, port,
		}
	}

	return &peersMap
}
