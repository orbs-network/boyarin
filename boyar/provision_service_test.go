package boyar

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_BoyarProvisionServices(t *testing.T) {
	orchestrator := &adapter.OrchestratorMock{}

	source := getJSONConfig(t, ConfigWithSigner)

	cache := NewCache()
	b := NewBoyar(orchestrator, source, cache, helpers.DefaultTestLogger())

	orchestrator.On("GetOverlayNetwork", mock.Anything, mock.Anything).Return("fake-network-id", nil)
	orchestrator.On("RunService", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionServices(context.Background())

	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func Test_BoyarSignerOffOn(t *testing.T) {
	orchestrator := &adapter.OrchestratorMock{}

	cache := NewCache()

	source := getJSONConfig(t, Config)
	require.EqualValues(t, "experimental", source.Services().Signer.DockerConfig.Tag)
	boyarWithoutSigner := NewBoyar(orchestrator, source, cache, helpers.DefaultTestLogger())

	orchestrator.On("GetOverlayNetwork", mock.Anything, mock.Anything, mock.Anything).Return("fake-network-id", nil).Once()
	orchestrator.On("RunService", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	err := boyarWithoutSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)

	sourceWithUpdatedSigner := getJSONConfig(t, ConfigWithSigner)
	require.EqualValues(t, "another-tag", sourceWithUpdatedSigner.Services().Signer.DockerConfig.Tag)

	orchestrator.On("GetOverlayNetwork", mock.Anything, mock.Anything, mock.Anything).Return("fake-network-id", nil).Once()
	orchestrator.On("RunService", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	boyarWithUpdatedSigner := NewBoyar(orchestrator, sourceWithUpdatedSigner, cache, helpers.DefaultTestLogger())
	err = boyarWithUpdatedSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

// Shouldn't be able to disable services, really
//func Test_BoyarSignerOnOff(t *testing.T) {
//	orchestrator := &adapter.OrchestratorMock{}
//
//	cache := NewCache()
//
//	sourceWithSigner := getJSONConfig(t, ConfigWithSigner)
//
//	orchestrator.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//	orchestrator.On("UpdateService", mock.Anything, mock.Anything).Return(nil).Once()
//
//	boyarWithSigner := NewBoyar(orchestrator, sourceWithSigner, cache, helpers.DefaultTestLogger())
//
//	err := boyarWithSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//
//	sourceWithoutSigner := getJSONConfig(t, Config)
//	boyarWithoutSigner := NewBoyar(orchestrator, sourceWithoutSigner, cache, helpers.DefaultTestLogger())
//
//	orchestrator.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//
//	err = boyarWithoutSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t) // nothing happens
//}
