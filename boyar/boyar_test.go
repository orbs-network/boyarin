package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"testing"
)

const input = `
{
	"keys": {
		"node-public-key": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
		"node-private-key": "93e919986a22477fda016789cca30cb841a135650938714f85f0000a65076bd4dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
		"constant-consensus-leader": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173"
	},
	"network": [
		{"Key":"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173","IP":"192.168.1.14"}
	],
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

	require.EqualValues(t, []*strelets.FederationNode{
		{
			Key: "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
			IP:  "192.168.1.14",
		},
	}, source.FederationNodes())

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

func TestNewUrlConfigurationSource(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	router := http.NewServeMux()
	router.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(input))
	})

	server := &http.Server{
		Handler: router,
	}

	go server.Serve(listener)
	defer server.Shutdown(context.TODO())

	source, err := NewUrlConfigurationSource(fmt.Sprintf("http://127.0.0.1:%d", listener.Addr().(*net.TCPAddr).Port))

	require.NoError(t, err)
	verifySource(t, source)
}
