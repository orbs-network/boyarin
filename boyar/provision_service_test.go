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
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(orchestrator, source, cache, helpers.DefaultTestLogger())

	orchestrator.On("GetOverlayNetwork", mock.Anything, mock.Anything).Return("fake-network-id", nil)
	orchestrator.On("RunService", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionServices(context.Background())

	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

//func Test_BoyarSignerOffOn(t *testing.T) {
//	streletsMock := &StreletsMock{}
//
//	cache := NewCache()
//
//	sourceWithoutSigner := getJSONConfig(t, Config)
//	sourceWithoutSigner.SetKeyConfigPath(fakeKeyPairPath)
//
//	boyarWithoutSigner := NewBoyar(streletsMock, sourceWithoutSigner, cache, helpers.DefaultTestLogger())
//
//	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//
//	err := boyarWithoutSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	streletsMock.AssertExpectations(t) // nothing happens
//
//	sourceWithSigner := getJSONConfig(t, ConfigWithSigner)
//	sourceWithSigner.SetKeyConfigPath(fakeKeyPairPath)
//	require.NoError(t, err)
//
//	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil).Once()
//
//	boyarWithSigner := NewBoyar(streletsMock, sourceWithSigner, cache, helpers.DefaultTestLogger())
//
//	err = boyarWithSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	streletsMock.AssertExpectations(t)
//
//	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//
//	err = boyarWithSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	streletsMock.AssertExpectations(t)
//}
//
//func Test_BoyarSignerOnOff(t *testing.T) {
//	streletsMock := &StreletsMock{}
//
//	cache := NewCache()
//
//	sourceWithSigner := getJSONConfig(t, ConfigWithSigner)
//	sourceWithSigner.SetKeyConfigPath(fakeKeyPairPath)
//
//	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil).Once()
//
//	boyarWithSigner := NewBoyar(streletsMock, sourceWithSigner, cache, helpers.DefaultTestLogger())
//
//	err := boyarWithSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	streletsMock.AssertExpectations(t)
//
//	sourceWithoutSigner := getJSONConfig(t, Config)
//	sourceWithoutSigner.SetKeyConfigPath(fakeKeyPairPath)
//
//	boyarWithoutSigner := NewBoyar(streletsMock, sourceWithoutSigner, cache, helpers.DefaultTestLogger())
//
//	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
//
//	err = boyarWithoutSigner.ProvisionServices(context.Background())
//	require.NoError(t, err)
//	streletsMock.AssertExpectations(t) // nothing happens
//}
