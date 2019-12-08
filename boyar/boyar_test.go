package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const fakeKeyPairPath = "./test/fake-key-pair.json"
const configPath = "./config/test"

func Test_BoyarProvisionVirtualChains(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := helpers.GetJSONConfig(t, configPath, helpers.Config)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Twice().Return(nil)
	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithErrors(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := helpers.GetJSONConfig(t, configPath, helpers.Config)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
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

	source := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithActiveVchains)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
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
	cfg := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()

	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.NotEmpty(t, cache.Get("42"), "cache should not be empty")

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	cfg := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.NotEmpty(t, cache.Get("42"), "cache should not be empty")

	cfg.Chains()[0].Config["active-consensus-algo"] = 999

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfDockerConfigChanges(t *testing.T) {
	cfg := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithActiveVchains)
	cfg.SetKeyConfigPath(fakeKeyPairPath)

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.NotEmpty(t, cache.Get("42"), "cache should not be empty")

	cfg.Chains()[1].DockerConfig.Tag = "beta"

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
	cfg := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithActiveVchains)

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "67565611bd34568f0b3ca12a8b257015701b3740d5e61300f19c3caa945db04a", cache.Get(config.HTTP_REVERSE_PROXY_HASH))

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
}

func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
	cfg := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithActiveVchains)

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "67565611bd34568f0b3ca12a8b257015701b3740d5e61300f19c3caa945db04a", cache.Get(config.HTTP_REVERSE_PROXY_HASH))

	cfg.Chains()[0].HttpPort = 9125

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 2)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 2)
}

func Test_BoyarProvisionVirtualChainsReprovisionsWithErrors(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithSingleChain)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	err := b.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "failed to provision virtual chain 42")
	require.Empty(t, cache.Get("1991"), "cache should be empty")
	streletsMock.VerifyMocks(t)

	streletsMockWithNoErrors := &StreletsMock{}
	streletsMockWithNoErrors.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	bWithNoErrors := NewBoyar(streletsMockWithNoErrors, source, cache, helpers.DefaultTestLogger())
	err = bWithNoErrors.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, cache.Get("42"), "cache should not be empty")
	streletsMockWithNoErrors.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsClearsCacheAfterFailedAttempts(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithSingleChain)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
	require.NotEmpty(t, cache.Get("42"), "cache should not be empty")

	streletsWithError := &StreletsMock{}
	streletsWithError.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	source.Chains()[0].DockerConfig.Tag = "new-tag"

	bWithError := NewBoyar(streletsWithError, source, cache, helpers.DefaultTestLogger())
	err = bWithError.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "failed to provision virtual chain 42")
	streletsWithError.VerifyMocks(t)
	require.Empty(t, cache.Get("42"), "should clear cache after failed attempt")
}

func Test_BoyarProvisionVirtualChainsClearsCacheAfterRemovingChain(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithSingleChain)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, cache.Get("42"), "cache should not be empty")
	streletsMock.VerifyMocks(t)

	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	source.Chains()[0].Disabled = true
	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	require.Empty(t, cache.Get("42"), "should clear cache")

	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionServices(t *testing.T) {
	streletsMock := &StreletsMock{}

	source := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithSigner)
	source.SetKeyConfigPath(fakeKeyPairPath)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil)
	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil)

	err := b.ProvisionServices(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarSignerOffOn(t *testing.T) {
	streletsMock := &StreletsMock{}

	cache := config.NewCache()

	sourceWithoutSigner := helpers.GetJSONConfig(t, configPath, helpers.Config)
	sourceWithoutSigner.SetKeyConfigPath(fakeKeyPairPath)

	boyarWithoutSigner := NewBoyar(streletsMock, sourceWithoutSigner, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()

	err := boyarWithoutSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t) // nothing happens

	sourceWithSigner := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithSigner)
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

	cache := config.NewCache()

	sourceWithSigner := helpers.GetJSONConfig(t, configPath, helpers.ConfigWithSigner)
	sourceWithSigner.SetKeyConfigPath(fakeKeyPairPath)

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()
	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil).Once()

	boyarWithSigner := NewBoyar(streletsMock, sourceWithSigner, cache, helpers.DefaultTestLogger())

	err := boyarWithSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)

	sourceWithoutSigner := helpers.GetJSONConfig(t, configPath, helpers.Config)
	sourceWithoutSigner.SetKeyConfigPath(fakeKeyPairPath)

	boyarWithoutSigner := NewBoyar(streletsMock, sourceWithoutSigner, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionSharedNetwork", mock.Anything, mock.Anything).Return(nil).Once()

	err = boyarWithoutSigner.ProvisionServices(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t) // nothing happens
}
