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

	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(streletsMock, source, cache, "/tmp/fake-key-pair.json")

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Twice()

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
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(s, config, cache, "./config.json")

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "4d775e0cd37e6c71e4aa4e0329fa56f8c47141ba202a8e900c5c46b05740e83d", cache["42"])

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(s, config, cache, "./config.json")

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "4d775e0cd37e6c71e4aa4e0329fa56f8c47141ba202a8e900c5c46b05740e83d", cache["42"])

	config.Chains()[0].Config["active-consensus-algo"] = 999

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfDockerConfigChanges(t *testing.T) {
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(s, config, cache, "./config.json")

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "4d775e0cd37e6c71e4aa4e0329fa56f8c47141ba202a8e900c5c46b05740e83d", cache["42"])

	config.Chains()[1].DockerConfig.Tag = "beta"

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(s, config, cache, "./config.json")

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "c8a7873c3324a608d8290a24e3a5168950a9588ef6b288043596e09f1977d058", cache[boyar.HTTP_REVERSE_PROXY_HASH])

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
}

func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
	config, _ := boyar.NewStringConfigurationSource(getJSONConfig())

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(boyar.BoyarConfigCache)
	b := boyar.NewBoyar(s, config, cache, "./config.json")

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "c8a7873c3324a608d8290a24e3a5168950a9588ef6b288043596e09f1977d058", cache[boyar.HTTP_REVERSE_PROXY_HASH])

	config.Chains()[0].HttpPort = 9125

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 2)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 2)
}
