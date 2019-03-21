package config

import (
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func getJSONConfig() string {
	contents, err := ioutil.ReadFile("./test/config.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func verifySource(t *testing.T, source NodeConfiguration) {
	require.EqualValues(t, []*strelets.FederationNode{
		{
			Address: "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
			IP:      "192.168.1.14",
		},
	}, source.FederationNodes())

	require.EqualValues(t, 3, len(source.Chains()))

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
	source, err := parseStringConfig(getJSONConfig())

	require.NoError(t, err)
	verifySource(t, source)
}

func TestNewStringConfigurationSource(t *testing.T) {
	source, err := NewStringConfigurationSource(getJSONConfig())

	require.NoError(t, err)
	verifySource(t, source)
}

func TestNewUrlConfigurationSource(t *testing.T) {
	server := helpers.CreateHttpServer("/", 0, func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(getJSONConfig()))
	})
	server.Start()
	defer server.Shutdown()

	source, err := NewUrlConfigurationSource(server.Url())

	require.NoError(t, err)
	verifySource(t, source)
}
