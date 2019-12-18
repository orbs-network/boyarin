package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"time"
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
	streletsMock := &StreletsMock{}

	source := getJSONConfig(t, Config)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Twice().Return(nil)
	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithErrors(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := getJSONConfig(t, Config)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(1991)
	})).Return(fmt.Errorf("unbearable catastrophe"))

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(42)
	})).Once().Return(nil)

	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())

	require.EqualError(t, err, "failed to provision virtual chain 1991")
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithTimeout(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := getJSONConfig(t, ConfigWithActiveVchains)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(1991)
	})).After(1 * time.Hour)

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(42)
	})).Once().Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := b.ProvisionVirtualChains(ctx)
	require.EqualError(t, err, "failed to provision virtual chain context deadline exceeded")
	streletsMock.VerifyMocks(t)
}

func TestBoyar_ProvisionVirtualChainsWithNoConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := NewCache()

	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	assertAllChainedCached(t, cfg, cache)

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	assertAllChainedCached(t, cfg, cache)

	cfg.Chains()[0].Config["active-consensus-algo"] = 999

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfDockerConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	assertAllChainedCached(t, cfg, cache)

	cfg.Chains()[1].DockerConfig.Tag = "beta"

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)

	assert.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(cfg)))

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
}

func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
	cfg := getJSONConfig(t, ConfigWithActiveVchains)

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)

	assert.False(t, cache.nginx.CheckNewJsonValue(getNginxConfig(cfg)))

	cfg.Chains()[0].HttpPort = 9125

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 2)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 2)
}

func Test_BoyarProvisionVirtualChainsReprovisionsWithErrors(t *testing.T) {
	streletsMock := &StreletsMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, cfg, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	err := b.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "failed to provision virtual chain 42")
	streletsMock.VerifyMocks(t)

	streletsMockWithNoErrors := &StreletsMock{}
	streletsMockWithNoErrors.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	bWithNoErrors := NewBoyar(streletsMockWithNoErrors, cfg, cache, helpers.DefaultTestLogger())
	err = bWithNoErrors.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)
	streletsMockWithNoErrors.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsClearsCacheAfterFailedAttempts(t *testing.T) {
	streletsMock := &StreletsMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, cfg, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
	assertAllChainedCached(t, cfg, cache)

	streletsWithError := &StreletsMock{}
	streletsWithError.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	cfg.Chains()[0].DockerConfig.Tag = "new-tag"

	bWithError := NewBoyar(streletsWithError, cfg, cache, helpers.DefaultTestLogger())
	err = bWithError.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "failed to provision virtual chain 42")
	streletsWithError.VerifyMocks(t)
	assert.True(t, cache.vChains.CheckNewJsonValue(cfg.Chains()[0].Id.String(), getVirtualChainConfig(cfg, cfg.Chains()[0])), "cache should not remember chain deployed with configuration")
}

func Test_BoyarProvisionVirtualChainsUpdatesCacheAfterRemovingChain(t *testing.T) {
	streletsMock := &StreletsMock{}

	cfg := getJSONConfig(t, ConfigWithSingleChain)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, cfg, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)
	streletsMock.VerifyMocks(t)

	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	cfg.Chains()[0].Disabled = true
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	assertAllChainedCached(t, cfg, cache)

	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionServices(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := getJSONConfig(t, ConfigWithSigner)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil)
	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionServices(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarSignerOffOn(t *testing.T) {
	streletsMock := &StreletsMock{}

	cache := NewCache()

	sourceWithoutSigner := getJSONConfig(t, Config)
	sourceWithoutSigner.SetKeyConfigPath(fakeKeyPairPath)

	boyarWithoutSigner := NewBoyar(streletsMock, sourceWithoutSigner, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()

	err := boyarWithoutSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t) // nothing happens

	sourceWithSigner := getJSONConfig(t, ConfigWithSigner)
	sourceWithSigner.SetKeyConfigPath(fakeKeyPairPath)
	require.NoError(t, err)

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil).Once()

	boyarWithSigner := NewBoyar(streletsMock, sourceWithSigner, cache, helpers.DefaultTestLogger())

	err = boyarWithSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()

	err = boyarWithSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarSignerOnOff(t *testing.T) {
	streletsMock := &StreletsMock{}

	cache := NewCache()

	sourceWithSigner := getJSONConfig(t, ConfigWithSigner)
	sourceWithSigner.SetKeyConfigPath(fakeKeyPairPath)

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil).Once()

	boyarWithSigner := NewBoyar(streletsMock, sourceWithSigner, cache, helpers.DefaultTestLogger())

	err := boyarWithSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)

	sourceWithoutSigner := getJSONConfig(t, Config)
	sourceWithoutSigner.SetKeyConfigPath(fakeKeyPairPath)

	boyarWithoutSigner := NewBoyar(streletsMock, sourceWithoutSigner, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()

	err = boyarWithoutSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t) // nothing happens
}
