package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getNetworkConfigJSON(t *testing.T) {
	var nodes []*config.FederationNode

	for i, address := range helpers.NodeAddresses() {
		nodes = append(nodes, &config.FederationNode{
			Address: address,
			IP:      fmt.Sprintf("10.0.0.%d", i+1),
			Port:    4400 + i,
		})
	}

	require.JSONEq(t, `{
		"federation-nodes": [
			{"address":"d27e2e7398e2582f63d0800330010b3e58952ff6","ip":"10.0.0.2","port":4401},
			{"address":"c056dfc0d1fbc7479db11e61d1b0b57612bf7f17", "ip":"10.0.0.4", "port":4403}, 
			{"address":"a328846cd5b4979d68a8c58a9bdfeee657b34de7","ip":"10.0.0.1","port":4400},
			{"address":"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9","ip":"10.0.0.3","port":4402}
		]
	}`, string(getNetworkConfigJSON(nodes)))
}

func Test_BoyarProvisionVirtualChains(t *testing.T) {
	orchestrator := &adapter.OrchestratorMock{}

	source := getJSONConfig(t, Config)

	cache := NewCache()
	b := NewBoyar(orchestrator, source, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil)
	orchestrator.On("RemoveService", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func TestBoyar_ProvisionVirtualChainsWithNoConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator := &adapter.OrchestratorMock{}
	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
	assertAllChainedCached(t, cfg, cache)

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator := &adapter.OrchestratorMock{}
	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
	assertAllChainedCached(t, cfg, cache)

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	cfg.Chains()[0].Config["active-consensus-algo"] = 999

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfDockerConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator := &adapter.OrchestratorMock{}
	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Twice()

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
	assertAllChainedCached(t, cfg, cache)

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	cfg.Chains()[1].DockerConfig.Tag = "beta"

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func Test_BoyarProvisionVirtualChainsReprovisionsWithErrors(t *testing.T) {
	orchestrator := &adapter.OrchestratorMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe")).Once()
	err := b.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "unbearable catastrophe")
	orchestrator.AssertExpectations(t)

	orchestratorWithNoErrors := &adapter.OrchestratorMock{}
	orchestratorWithNoErrors.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	bWithNoErrors := NewBoyar(orchestratorWithNoErrors, cfg, cache, helpers.DefaultTestLogger())
	err = bWithNoErrors.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)
	orchestratorWithNoErrors.AssertExpectations(t)
}

func Test_BoyarProvisionVirtualChainsClearsCacheAfterFailedAttempts(t *testing.T) {
	orchestrator := &adapter.OrchestratorMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
	assertAllChainedCached(t, cfg, cache)

	orchestratorWithError := &adapter.OrchestratorMock{}
	orchestratorWithError.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	cfg.Chains()[0].DockerConfig.Tag = "new-tag"

	bWithError := NewBoyar(orchestratorWithError, cfg, cache, helpers.DefaultTestLogger())
	err = bWithError.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "unbearable catastrophe")
	orchestratorWithError.AssertExpectations(t)
	assert.True(t, cache.vChains.CheckNewJsonValue(cfg.Chains()[0].Id.String(), getVirtualChainConfig(cfg, cfg.Chains()[0])), "cache should not remember chain deployed with configuration")
}

func Test_BoyarProvisionVirtualChainsOnOffAndOnAgain(t *testing.T) {
	orchestrator := &adapter.OrchestratorMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)
	orchestrator.AssertExpectations(t)

	orchestrator.On("RemoveService", mock.Anything, mock.Anything).Return(nil)

	cfg.Chains()[0].Disabled = true
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)

	orchestrator.AssertExpectations(t)

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	cfg.Chains()[0].Disabled = false
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)

	orchestrator.AssertExpectations(t)
}
