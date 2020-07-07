package config

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
	"time"
)

func getTestJSONConfig() string {
	contents, err := ioutil.ReadFile("./test/config.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

const fakeKeyPair = "./test/fake-key-pair.json"

func Test_StringConfigurationSource(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "", fakeKeyPair, false)
	require.NoError(t, err)

	require.NotEmpty(t, source.Hash())

	require.Equal(t, "http://localhost:8545", source.Chains()[0].Config["ethereum-endpoint"])

	require.False(t, source.Chains()[0].Disabled)
	require.False(t, source.Chains()[1].Disabled)
	require.True(t, source.Chains()[2].Disabled)

	require.NotNil(t, source.OrchestratorOptions())
	require.Equal(t, "ebs", source.OrchestratorOptions().StorageDriver)
	require.NotNil(t, source.OrchestratorOptions().StorageOptions)
	require.Equal(t, "100", source.OrchestratorOptions().StorageOptions["maxRetries"])

	resources := source.Chains()[0].DockerConfig.Resources
	require.EqualValues(t, 1024, resources.Limits.Memory)
	require.EqualValues(t, 1, resources.Limits.CPUs)
	require.EqualValues(t, 512, resources.Reservations.Memory)
	require.EqualValues(t, 0.5, resources.Reservations.CPUs)
	require.EqualValues(t, 5, source.Chains()[0].DockerConfig.Volumes.Blocks)
	require.EqualValues(t, 1, source.Chains()[0].DockerConfig.Volumes.Logs)

	require.NotNil(t, source, source.Chains()[1].DockerConfig.Resources.Limits)
	require.NotNil(t, source, source.Chains()[1].DockerConfig.Resources.Reservations)

	require.NotNil(t, source.Services())
	require.NotEmpty(t, source.Chains()[0].Config["signer-endpoint"])

	require.EqualValues(t, 8080, source.Chains()[0].InternalHttpPort)
}

func Test_StringConfigurationSourceFromEmptyConfig(t *testing.T) {
	cfg, err := NewStringConfigurationSource("{}", "", fakeKeyPair, false)
	require.NoError(t, err)

	require.NotEmpty(t, cfg.Hash())
	require.Empty(t, cfg.Chains())
	require.Empty(t, cfg.FederationNodes())
	require.NotNil(t, cfg.OrchestratorOptions())
}

func Test_StringConfigurationSourceWithOverrides(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, false)
	require.NoError(t, err)

	require.Equal(t, "http://some.ethereum.node", source.EthereumEndpoint())
	require.Equal(t, 1*time.Minute, source.OrchestratorOptions().MaxReloadTimedDelay())
	require.Equal(t, "http://some.ethereum.node", source.Chains()[0].Config["ethereum-endpoint"])
}

func Test_StringConfigurationSourceWithSigner(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, false)
	require.NoError(t, err)

	require.NotNil(t, source.Services())
	require.NotNil(t, source.Services().Signer())
	require.NotNil(t, source.Services().Signer().DockerConfig)
	require.NotNil(t, source.Services().Signer().Config)

	require.Equal(t, "http://signer:7777", source.Chains()[0].Config["signer-endpoint"])
}

func Test_StringConfigurationSourceWithSignerWithNamespace(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, true)
	require.NoError(t, err)

	require.NotNil(t, source.Services())
	require.NotNil(t, source.Services().Signer())
	require.NotNil(t, source.Services().Signer().DockerConfig)
	require.NotNil(t, source.Services().Signer().Config)

	require.Equal(t, "http://cfc9e5-signer:7777", source.Chains()[0].Config["signer-endpoint"])
}

func Test_StringConfigurationSourceWithCustomService(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, false)
	require.NoError(t, err)

	require.NotNil(t, source.Services())
	signer := source.Services().Signer()
	require.NotNil(t, signer)
	require.NotNil(t, signer.DockerConfig)
	require.NotNil(t, signer.Config)

	require.True(t, signer.AllowAccessToSigner)
	require.False(t, signer.AllowAccessToServices)
	require.EqualValues(t, "/opt/orbs/orbs-signer", signer.ExecutablePath)

	require.Equal(t, "http://signer:7777", source.Chains()[0].Config["signer-endpoint"])

	customService := source.Services()["service-name"]
	require.NotNil(t, customService)
	require.NotNil(t, customService.Config)
	require.NotNil(t, customService.DockerConfig)

	require.False(t, customService.AllowAccessToSigner)
	require.True(t, customService.AllowAccessToServices)
	require.EqualValues(t, "/opt/orbs/service", customService.ExecutablePath)
}
