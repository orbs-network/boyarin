package boyar

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

const input = `
[
	{
		"keys": {
			"node-public-key": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
			"node-private-key": "93e919986a22477fda016789cca30cb841a135650938714f85f0000a65076bd4dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
			"constant-consensus-leader": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173"
		},
		"vchain": {
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
	}
]
`

func Test_BootstrapFromJSON(t *testing.T) {
	source, err := parseStringConfig(input)

	require.NoError(t, err)
	require.EqualValues(t, 1, len(source.values))

	fmt.Println(fmt.Sprintf("%v", source.values[0]))

	chain := source.values[0].vchain
	require.EqualValues(t, 42, chain.Id)
}
