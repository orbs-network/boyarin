package boyar

import (
	"github.com/stretchr/testify/require"
	"testing"
)

const input = `
{
	"keys": {
		"node-public-key": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
		"node-private-key": "93e919986a22477fda016789cca30cb841a135650938714f85f0000a65076bd4dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
		"constant-consensus-leader": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173"
	},
	"chains": [
		{
			"Id":        42,
			"HttpPort":   8080,
			"GossipPort": 4400,
			"DockerConfig": {
			"ContainerNamePrefix": "node1",
				"Image":  "506367651493.dkr.ecr.us-west-2.amazonaws.com/orbs-network-v1",
				"Tag":    "master",
				"Pull":   false
			}
		}
	]
}
`

func verifySource(t *testing.T, source ConfigurationSource) {
	require.NotEqual(t, []byte("null"), source.Keys())
	require.EqualValues(t, 1, len(source.Chains()))

	chain := source.Chains()[0]

	require.EqualValues(t, 42, chain.Id)
	require.EqualValues(t, 8080, chain.HttpPort)
	require.EqualValues(t, 4400, chain.GossipPort)

	require.EqualValues(t, "node1", chain.DockerConfig.ContainerNamePrefix)
	require.EqualValues(t, "506367651493.dkr.ecr.us-west-2.amazonaws.com/orbs-network-v1", chain.DockerConfig.Image)
	require.EqualValues(t, "master", chain.DockerConfig.Tag)
	require.EqualValues(t, false, chain.DockerConfig.Pull)
}

func Test_parseStringConfig(t *testing.T) {
	source, err := parseStringConfig(input)

	require.NoError(t, err)
	verifySource(t, source)
}

func TestNewStringConfigurationSource(t *testing.T) {
	source, err := NewStringConfigurationSource(input)

	require.NoError(t, err)
	verifySource(t, source)
}
