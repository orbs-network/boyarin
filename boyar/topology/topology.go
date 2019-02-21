package topology

import (
	"github.com/orbs-network/boyarin/boyar/topology/ethereum"
	"github.com/orbs-network/boyarin/strelets"
	"net"
	"strings"
)

func IpToString(ip [4]byte) string {
	return net.IPv4(ip[0], ip[1], ip[2], ip[3]).String()
}

func EthereumToOrbsAddress(eth string) string {
	return strings.ToLower(eth[2:])
}

func Convert(rawTopology *ethereum.RawTopology) (federationNodes []*strelets.FederationNode) {
	for index, address := range rawTopology.NodeAddresses {
		federationNodes = append(federationNodes, &strelets.FederationNode{
			Key: EthereumToOrbsAddress(address.Hex()),
			IP:  IpToString(rawTopology.IpAddresses[index]),
		})
	}

	return
}
