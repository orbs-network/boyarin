package strelets

import (
	"fmt"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNetworkConfigJSON(t *testing.T) {
	peers := make(PeersMap)

	for i, key := range test.NodeAddresses() {
		peers[NodeAddress(key)] = &Peer{
			IP:   fmt.Sprintf("10.0.0.%d", i+1),
			Port: 4400 + i,
		}
	}

	require.JSONEq(t, `{
		"federation-nodes": [
			{"address":"d27e2e7398e2582f63d0800330010b3e58952ff6","ip":"10.0.0.2","port":4401},
			{"address":"a328846cd5b4979d68a8c58a9bdfeee657b34de7","ip":"10.0.0.1","port":4400},
			{"address":"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9","ip":"10.0.0.3","port":4402}
		]
	}`, string(getNetworkConfigJSON(&peers)))
}
