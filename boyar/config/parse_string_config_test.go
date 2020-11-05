package config

import (
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
	require.EqualValues(t, []*FederationNode{
		{
			Address: "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
			IP:      "192.168.1.14",
		},
	}, source.FederationNodes())

	require.EqualValues(t, 3, len(source.Chains()))

	chain := source.Chains()[0]

	require.EqualValues(t, 42, chain.Id)
	require.EqualValues(t, 4400, chain.InternalPort)

	require.EqualValues(t, "orbsnetwork/node", chain.DockerConfig.Image)
	require.EqualValues(t, "experimental", chain.DockerConfig.Tag)
	require.EqualValues(t, false, chain.DockerConfig.Pull)
}

func Test_parseStringConfig(t *testing.T) {
	source, err := parseStringConfig(getJSONConfig(), "", fakeKeyPair, false)

	require.NoError(t, err)
	verifySource(t, source)
}

func TestNewStringConfigurationSource(t *testing.T) {
	source, err := NewStringConfigurationSource(getJSONConfig(), "", fakeKeyPair, false)

	require.NoError(t, err)
	verifySource(t, source)
}

func TestNewUrlConfigurationSource(t *testing.T) {
	server := helpers.CreateHttpServer("/", 0, func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte(getJSONConfig()))
	})
	server.Start()
	defer server.Shutdown()

	source, err := NewUrlConfigurationSource(server.Url(), "", fakeKeyPair, false)

	require.NoError(t, err)
	verifySource(t, source)
}

func TestNewUrlConfigurationSourceWithFaultyConfig(t *testing.T) {
	server := helpers.CreateHttpServer("/", 0, func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusInternalServerError)
	})
	server.Start()
	defer server.Shutdown()

	source, err := NewUrlConfigurationSource(server.Url(), "", fakeKeyPair, false)

	require.EqualError(t, err, "management config url returned with status 500 Internal Server Error")
	require.Nil(t, source)
}
