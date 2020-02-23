package boyar

import (
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/boyar/topology"
)

func buildPeersMap(nodes []*topology.FederationNode, gossipPort int) *config.PeersMap {
	peersMap := make(config.PeersMap)

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		peersMap[config.NodeAddress(node.Address)] = &config.Peer{
			node.IP, port,
		}
	}

	return &peersMap
}

func getKeyConfigJson(config config.NodeConfiguration, addressOnly bool) []byte {
	keyConfig := config.KeyConfig()
	if keyConfig == nil {
		return []byte{}
	}
	return keyConfig.JSON(addressOnly)
}
