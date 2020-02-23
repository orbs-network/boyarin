package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	. "github.com/orbs-network/boyarin/test"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

const fakeKeyPairPath = "./test/fake-key-pair.json"
const configPath = "./config/test"

type configFile int

const (
	Config configFile = iota
	ConfigWithActiveVchains
	ConfigWithSingleChain
	ConfigWithSigner
)

func (conf configFile) String() string {
	switch conf {
	case Config:
		return "config.json"
	case ConfigWithActiveVchains:
		return "configWithActiveVchains.json"
	case ConfigWithSingleChain:
		return "configWithSingleChain.json"
	case ConfigWithSigner:
		return "configWithSigner.json"
	default:
		panic(fmt.Sprintf("unknown config: %d", conf))
	}
}

func getJSONConfig(t *testing.T, conf configFile) config.MutableNodeConfiguration {
	contents, err := ioutil.ReadFile(configPath + "/" + conf.String())
	require.NoError(t, err)
	source, err := config.NewStringConfigurationSource(string(contents), helpers.LocalEthEndpoint()) // ethereum endpoint is optional
	require.NoError(t, err)
	return source
}

func assertAllChainedCached(t *testing.T, cfg config.MutableNodeConfiguration, cache *Cache) {
	for _, chain := range cfg.Chains() {
		if chain.Disabled {
			assert.False(t, cache.vChains.CheckNewValue(chain.Id.String(), removed), "cache should remember chain was removed")
		} else {
			assert.False(t, cache.vChains.CheckNewJsonValue(chain.Id.String(), getVirtualChainConfig(cfg, chain)), "cache should remember chain deployed with configuration")
		}
	}
}

func Test_BoyarProvisionVirtualChains(t *testing.T) {
	orchestrator := &OrchestratorMock{}

	source := getJSONConfig(t, Config)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(orchestrator, source, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Twice().Return(nil)
	orchestrator.On("ServiceRemove", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
}

func TestBoyar_ProvisionVirtualChainsWithNoConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator := &OrchestratorMock{}
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
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator := &OrchestratorMock{}
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
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator := &OrchestratorMock{}
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

//func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
//	cfg := getJSONConfig(t, ConfigWithActiveVchains)
//
//	orchestrator := &OrchestratorMock{}
//	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
//
//	cache := NewCache()
//	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())
//
//	err := b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//
//	nginxConfig := getNginxCompositeConfig(cfg)
//	assert.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(nginxConfig.Chains, nginxConfig.IP, false)))
//
//	err = b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//}

//func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
//	cfg := getJSONConfig(t, ConfigWithActiveVchains)
//
//	orchestrator := &OrchestratorMock{}
//	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
//
//	cache := NewCache()
//	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())
//
//	err := b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//
//	assert.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(cfg.Chains(), "127.0.0.1", false)))
//
//	orchestrator.On("RunReverseProxy", mock.Anything, mock.Anything).Return(nil).Once()
//	cfg.Chains()[0].HttpPort = 9125
//
//	err = b.ProvisionHttpAPIEndpoint(context.Background())
//	require.NoError(t, err)
//	orchestrator.AssertExpectations(t)
//}

func Test_BoyarProvisionVirtualChainsReprovisionsWithErrors(t *testing.T) {
	orchestrator := &OrchestratorMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe")).Once()
	err := b.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "unbearable catastrophe")
	orchestrator.AssertExpectations(t)

	orchestratorWithNoErrors := &OrchestratorMock{}
	orchestratorWithNoErrors.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	bWithNoErrors := NewBoyar(orchestratorWithNoErrors, cfg, cache, helpers.DefaultTestLogger())
	err = bWithNoErrors.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)
	orchestratorWithNoErrors.AssertExpectations(t)
}

func Test_BoyarProvisionVirtualChainsClearsCacheAfterFailedAttempts(t *testing.T) {
	orchestrator := &OrchestratorMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertExpectations(t)
	assertAllChainedCached(t, cfg, cache)

	orchestratorWithError := &OrchestratorMock{}
	orchestratorWithError.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	cfg.Chains()[0].DockerConfig.Tag = "new-tag"

	bWithError := NewBoyar(orchestratorWithError, cfg, cache, helpers.DefaultTestLogger())
	err = bWithError.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "unbearable catastrophe")
	orchestratorWithError.AssertExpectations(t)
	assert.True(t, cache.vChains.CheckNewJsonValue(cfg.Chains()[0].Id.String(), getVirtualChainConfig(cfg, cfg.Chains()[0])), "cache should not remember chain deployed with configuration")
}

func Test_BoyarProvisionVirtualChainsUpdatesCacheAfterRemovingChain(t *testing.T) {
	orchestrator := &OrchestratorMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(orchestrator, cfg, cache, helpers.DefaultTestLogger())

	orchestrator.On("RunVirtualChain", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)
	orchestrator.AssertExpectations(t)

	orchestrator.On("ServiceRemove", mock.Anything, mock.Anything).Return(nil)

	cfg.Chains()[0].Disabled = true
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)

	orchestrator.AssertExpectations(t)
}

func Test_BoyarProvisionServices(t *testing.T) {
	orchestrator := &OrchestratorMock{}

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
