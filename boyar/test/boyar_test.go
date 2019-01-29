package test

import (
	"context"
	"github.com/orbs-network/boyarin/boyar"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func getJSONConfig() string {
	contents, err := ioutil.ReadFile("./config.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func Test_BoyarProvisionVirtualChains(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := boyar.NewStringConfigurationSource(getJSONConfig())
	require.NoError(t, err)

	b := boyar.NewBoyar(streletsMock, source, "/tmp/fake-key-pair.json")

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Once()

	err = b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_StringConfigurationSource(t *testing.T) {
	source, err := boyar.NewStringConfigurationSource(getJSONConfig())
	require.NoError(t, err)

	require.NotEmpty(t, source.Hash())

	require.Equal(t, "http://localhost:8545", source.Chains()[0].Config["ethereum-endpoint"])

	require.NotNil(t, source.OrchestratorOptions())
	require.Equal(t, "ebs", source.OrchestratorOptions().StorageDriver)
	require.NotNil(t, source.OrchestratorOptions().StorageOptions)
	require.Equal(t, "100", source.OrchestratorOptions().StorageOptions["maxRetries"])
}

func Test_StringConfigurationSourceFromEmptyConfig(t *testing.T) {
	source, err := boyar.NewStringConfigurationSource("{}")
	require.NoError(t, err)

	require.NotEmpty(t, source.Hash())
	require.Empty(t, source.Chains())
	require.Empty(t, source.FederationNodes())
	require.NotNil(t, source.OrchestratorOptions())
}

func TestBoyar_ProvisionVirtualChainsWithNoConfigChanges(t *testing.T) {
	t.Skip("temp skip")

	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator := &OrchestratorMock{}

	s := strelets.NewStrelets(orchestrator)
	b := boyar.NewBoyar(s, config, "./config.json")

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
}
