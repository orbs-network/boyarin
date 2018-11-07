package config

import (
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetProvisionVirtualChainInputFromJSON(t *testing.T) {
	args := []string{
		"--chain-config", "./_fixtures/config.json",
		"--keys-config", "../e2e-config/node1-keys.json",
		"--peers-config", "./_fixtures/network.json",
	}

	expectedPeers := make(strelets.PeersMap)
	expectedPeers["dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173"] = &strelets.Peer{
		IP:   "192.168.1.14",
		Port: 4401,
	}

	expectedInput := &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:         99,
			HttpPort:   8081,
			GossipPort: 4401,
			DockerConfig: &strelets.DockerImageConfig{
				ContainerNamePrefix: "node1",
				Image:               "orbs",
				Tag:                 "export",
				Pull:                false,
			},
		},
		KeyPairConfigPath: "../e2e-config/node1-keys.json",
		Peers:             &expectedPeers,
	}

	input, err := GetProvisionVirtualChainInput(args)
	require.NoError(t, err)

	require.EqualValues(t, expectedInput, input)
}
