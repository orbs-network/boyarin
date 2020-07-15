package boyar

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator := &adapter.OrchestratorMock{}
	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)

	require.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(cfg)))

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator := &adapter.OrchestratorMock{}
	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)

	require.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(cfg)))

	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
	cfg.Chains()[0].Id = 9125

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}
