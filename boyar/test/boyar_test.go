package test

import (
	"context"
	"github.com/orbs-network/boyarin/boyar"
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
	streletsMock := &streletsMock{
		mock: &mock.Mock{},
	}

	source, err := boyar.NewStringConfigurationSource(getJSONConfig())
	require.NoError(t, err)

	b := boyar.NewBoyar(streletsMock, source, "/tmp/fake-key-pair.json")

	streletsMock.mock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Once()

	err = b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_StringConfigurationSource(t *testing.T) {
	source, err := boyar.NewStringConfigurationSource(getJSONConfig())
	require.NoError(t, err)

	require.Equal(t, "http://localhost:8545", source.Chains()[0].Config["ethereum-endpoint"])

	require.NotNil(t, source.OrchestratorOptions())
	require.Equal(t, "ebs", source.OrchestratorOptions().StorageDriver())
	require.NotNil(t, source.OrchestratorOptions().StorageOptions())
	require.Equal(t, "100", source.OrchestratorOptions().StorageOptions()["maxRetries"])
}
