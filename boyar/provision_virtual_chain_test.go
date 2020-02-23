package boyar

import (
	"fmt"
	"github.com/orbs-network/boyarin/boyar/topology"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNetworkConfigJSON(t *testing.T) {
	var nodes []*topology.FederationNode

	for i, address := range helpers.NodeAddresses() {
		nodes = append(nodes, &topology.FederationNode{
			Address: address,
			IP:      fmt.Sprintf("10.0.0.%d", i+1),
			Port:    4400 + i,
		})
	}

	require.JSONEq(t, `{
		"federation-nodes": [
			{"address":"d27e2e7398e2582f63d0800330010b3e58952ff6","ip":"10.0.0.2","port":4401},
			{"address":"c056dfc0d1fbc7479db11e61d1b0b57612bf7f17", "ip":"10.0.0.4", "port":4403}, 
			{"address":"a328846cd5b4979d68a8c58a9bdfeee657b34de7","ip":"10.0.0.1","port":4400},
			{"address":"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9","ip":"10.0.0.3","port":4402}
		]
	}`, string(getNetworkConfigJSON(nodes)))
}
