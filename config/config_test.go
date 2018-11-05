package config

import (
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetProvisionVirtualChainInput(t *testing.T) {
	args := []string{
		"--chain-config", "./_fixtures/config.json",
		"--keys-config", "../e2e-config/node1-keys.json",
	}

	expectedInput := &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:         99,
			HttpPort:   8081,
			GossipPort: 4401,
			DockerConfig: &strelets.DockerImageConfig{
				Prefix: "node1",
				Image:  "orbs",
				Tag:    "export",
				Pull:   false,
			},
		},
		KeysConfigPath: "../e2e-config/node1-keys.json",
		Peers:          nil,
	}

	input, err := GetProvisionVirtualChainInput(args)
	require.NoError(t, err)

	require.EqualValues(t, expectedInput, input)
}
