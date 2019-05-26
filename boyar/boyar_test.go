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
	"io/ioutil"
	"testing"
	"time"
)

func getJSONConfig() string {
	contents, err := ioutil.ReadFile("./config/test/config.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func getJSONConfigWithSigner() string {
	contents, err := ioutil.ReadFile("./config/test/configWithSigner.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func getJSONConfigWithSingleChain() string {
	contents, err := ioutil.ReadFile("./config/test/configWithSingleChain.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func getJSONConfigWithActiveVchains() string {
	contents, err := ioutil.ReadFile("./config/test/configWithActiveVchains.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func Test_BoyarProvisionVirtualChains(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfig(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Twice().Return(nil)
	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err = b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithErrors(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfig(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(1991)
	})).Return(fmt.Errorf("unbearable catastrophe"))

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(42)
	})).Once().Return(nil)

	streletsMock.On("RemoveVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err = b.ProvisionVirtualChains(context.Background())

	require.EqualError(t, err, "failed to provision virtual chain 1991")
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithTimeout(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfigWithActiveVchains(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

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

	err = b.ProvisionVirtualChains(ctx)
	require.EqualError(t, err, "failed to provision virtual chain context deadline exceeded")
	streletsMock.VerifyMocks(t)
}

func TestBoyar_ProvisionVirtualChainsWithNoConfigChanges(t *testing.T) {
	cfg, err := config.NewStringConfigurationSource(getJSONConfigWithActiveVchains(), "")
	cfg.SetKeyConfigPath("./test/fake-key-pair.json")

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()

	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "525beda4c5da8fb73233768536fe283b53dfa4173f536ac9ad8fe5f57ad3a5c7", cache.Get("42"))

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfigWithActiveVchains(), "")
	cfg.SetKeyConfigPath("./test/fake-key-pair.json")

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "525beda4c5da8fb73233768536fe283b53dfa4173f536ac9ad8fe5f57ad3a5c7", cache.Get("42"))

	cfg.Chains()[0].Config["active-consensus-algo"] = 999

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfDockerConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfigWithActiveVchains(), "")
	cfg.SetKeyConfigPath("./test/fake-key-pair.json")

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := config.NewCache()
	b := NewBoyar(s, cfg, cache, helpers.DefaultTestLogger())

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "525beda4c5da8fb73233768536fe283b53dfa4173f536ac9ad8fe5f57ad3a5c7", cache.Get("42"))

	cfg.Chains()[1].DockerConfig.Tag = "beta"

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfigWithActiveVchains(), "")

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
	cfg, _ := config.NewStringConfigurationSource(getJSONConfigWithActiveVchains(), "")

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

	source, err := config.NewStringConfigurationSource(getJSONConfigWithSingleChain(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(fmt.Errorf("unbearable catastrophe"))
	err = b.ProvisionVirtualChains(context.Background())
	require.EqualError(t, err, "failed to provision virtual chain 42")
	require.Empty(t, cache.Get("1991"), "cache should be empty")
	streletsMock.VerifyMocks(t)

	streletsMockWithNoErrors := &StreletsMock{}
	streletsMockWithNoErrors.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	bWithNoErrors := NewBoyar(streletsMockWithNoErrors, source, cache, helpers.DefaultTestLogger())
	err = bWithNoErrors.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	require.Equal(t, "9f1ec56888af85f24bb844d75d99e1771a1dfd0103b11eb120f0c8bd6ced2a88", cache.Get("42"))
	streletsMockWithNoErrors.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsClearsCacheAfterFailedAttempts(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfigWithSingleChain(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
	require.Equal(t, "9f1ec56888af85f24bb844d75d99e1771a1dfd0103b11eb120f0c8bd6ced2a88", cache.Get("42"))

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

	source, err := config.NewStringConfigurationSource(getJSONConfigWithSingleChain(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Return(nil)

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	require.Equal(t, "9f1ec56888af85f24bb844d75d99e1771a1dfd0103b11eb120f0c8bd6ced2a88", cache.Get("42"))
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

	source, err := config.NewStringConfigurationSource(getJSONConfigWithSigner(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := config.NewCache()
	b := NewBoyar(streletsMock, source, cache, helpers.DefaultTestLogger())

	streletsMock.On("UpdateService", mock.Anything, mock.Anything).Return(nil)

	err = b.ProvisionServices(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}
