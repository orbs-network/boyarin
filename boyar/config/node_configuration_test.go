package config

import (
	"github.com/orbs-network/boyarin/boyar/topology"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeConfigurationContainer_Hash(t *testing.T) {
	source, err := parseStringConfig(getJSONConfig(), "", fakeKeyPair, false)
	require.NoError(t, err)

	oldHash := source.Hash()
	require.NotEmpty(t, oldHash, "hash can't be empty")

	require.EqualValues(t, oldHash, source.Hash(), "hash can't change if the value didn't change")

	source.value.FederationNodes = []*topology.FederationNode{
		{IP: "1.2.3.4", Address: "some-fake-key"},
	}

	require.NotEqual(t, oldHash, source.Hash(), "hash should have been changed")
}
