package boyar

import (
	"context"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets"
	. "github.com/orbs-network/boyarin/test"
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

func Test_BoyarProvisionVirtualChains(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfig())
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := make(config.BoyarConfigCache)
	b := NewBoyar(streletsMock, source, cache)

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.Anything).Twice().Return(nil)

	err = b.ProvisionVirtualChains(context.Background())

	require.NoError(t, err)
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithErrors(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfig())
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := make(config.BoyarConfigCache)
	b := NewBoyar(streletsMock, source, cache)

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(1991)
	})).Return(fmt.Errorf("unbearable catastrophe"))

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(42)
	})).Once().Return(nil)

	err = b.ProvisionVirtualChains(context.Background())

	require.EqualError(t, err, "failed to provision virtual chain 1991: unbearable catastrophe")
	streletsMock.VerifyMocks(t)
}

func Test_BoyarProvisionVirtualChainsWithTimeout(t *testing.T) {
	streletsMock := &StreletsMock{}

	source, err := config.NewStringConfigurationSource(getJSONConfig())
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	cache := make(config.BoyarConfigCache)
	b := NewBoyar(streletsMock, source, cache)

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(1991)
	})).After(1 * time.Hour)

	streletsMock.On("ProvisionVirtualChain", mock.Anything, mock.MatchedBy(func(input *strelets.ProvisionVirtualChainInput) bool {
		return input.VirtualChain.Id == strelets.VirtualChainId(42)
	})).Once().Return(nil)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = b.ProvisionVirtualChains(ctx)
	require.EqualError(t, err, "failed to provision virtual chain 1991: context deadline exceeded")
	streletsMock.VerifyMocks(t)
}

func TestBoyar_ProvisionVirtualChainsWithNoConfigChanges(t *testing.T) {
	cfg, err := config.NewStringConfigurationSource(getJSONConfig())
	cfg.SetKeyConfigPath("./test/fake-key-pair.json")

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(config.BoyarConfigCache)
	b := NewBoyar(s, cfg, cache)

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "f1fc45fe688c808324c1907ba7047c0dc7763ace14ee576e69ab9a56be7e55fc", cache["42"])

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfig())
	cfg.SetKeyConfigPath("./test/fake-key-pair.json")

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(config.BoyarConfigCache)
	b := NewBoyar(s, cfg, cache)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "f1fc45fe688c808324c1907ba7047c0dc7763ace14ee576e69ab9a56be7e55fc", cache["42"])

	cfg.Chains()[0].Config["active-consensus-algo"] = 999

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionVirtualChainsReprovisionsIfDockerConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfig())
	cfg.SetKeyConfigPath("./test/fake-key-pair.json")

	orchestrator, virtualChainRunner, _ := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(config.BoyarConfigCache)
	b := NewBoyar(s, cfg, cache)

	err := b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 2)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 2)
	require.EqualValues(t, "f1fc45fe688c808324c1907ba7047c0dc7763ace14ee576e69ab9a56be7e55fc", cache["42"])

	cfg.Chains()[1].DockerConfig.Tag = "beta"

	err = b.ProvisionVirtualChains(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "Prepare", 3)
	virtualChainRunner.AssertNumberOfCalls(t, "Run", 3)
}

func TestBoyar_ProvisionHttpAPIEndpointWithNoConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfig())

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(config.BoyarConfigCache)
	b := NewBoyar(s, cfg, cache)

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "c8a7873c3324a608d8290a24e3a5168950a9588ef6b288043596e09f1977d058", cache[config.HTTP_REVERSE_PROXY_HASH])

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
}

func TestBoyar_ProvisionHttpAPIEndpointReprovisionsIfConfigChanges(t *testing.T) {
	cfg, _ := config.NewStringConfigurationSource(getJSONConfig())

	orchestrator, _, httpProxyRunner := NewOrchestratorAndRunnerMocks()

	s := strelets.NewStrelets(orchestrator)
	cache := make(config.BoyarConfigCache)
	b := NewBoyar(s, cfg, cache)

	err := b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 1)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 1)
	require.EqualValues(t, "c8a7873c3324a608d8290a24e3a5168950a9588ef6b288043596e09f1977d058", cache[config.HTTP_REVERSE_PROXY_HASH])

	cfg.Chains()[0].HttpPort = 9125

	err = b.ProvisionHttpAPIEndpoint(context.Background())
	require.NoError(t, err)
	orchestrator.AssertNumberOfCalls(t, "PrepareReverseProxy", 2)
	httpProxyRunner.AssertNumberOfCalls(t, "Run", 2)
}
