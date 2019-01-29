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

	cache := make(boyar.BoyarConfigCache)
	_, err = b.ProvisionVirtualChains(context.Background(), cache)

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
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, runner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	b := boyar.NewBoyar(s, config, "./config.json")

	cache := make(boyar.BoyarConfigCache)

	cache, err := b.ProvisionVirtualChains(context.Background(), cache)
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 1)
	runner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "4d775e0cd37e6c71e4aa4e0329fa56f8c47141ba202a8e900c5c46b05740e83d", cache[strelets.VirtualChainId(42)])

	cache, err = b.ProvisionVirtualChains(context.Background(), cache)
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 1)
	runner.AssertNumberOfCalls(t, "Run", 1)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, runner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	b := boyar.NewBoyar(s, config, "./config.json")

	cache := make(boyar.BoyarConfigCache)

	cache, err := b.ProvisionVirtualChains(context.Background(), cache)
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 1)
	runner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "4d775e0cd37e6c71e4aa4e0329fa56f8c47141ba202a8e900c5c46b05740e83d", cache[strelets.VirtualChainId(42)])

	config.Chains()[0].Config["active-consensus-algo"] = 999

	cache, err = b.ProvisionVirtualChains(context.Background(), cache)
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	runner.AssertNumberOfCalls(t, "Run", 2)
}
