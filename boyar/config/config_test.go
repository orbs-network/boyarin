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

func getTestJSONConfigWithOverrides() string {
	contents, err := ioutil.ReadFile("./test/configWithOverrides.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func getTestJSONConfigSigner() string {
	contents, err := ioutil.ReadFile("./test/configWithSigner.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func Test_StringConfigurationSource(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "")
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
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

	require.EqualValues(t, 1024, source.Chains()[0].DockerConfig.Resources.Limits.Memory)
	require.EqualValues(t, 1, source.Chains()[0].DockerConfig.Resources.Limits.CPUs)
	require.EqualValues(t, 512, source.Chains()[0].DockerConfig.Resources.Reservations.Memory)
	require.EqualValues(t, 0.5, source.Chains()[0].DockerConfig.Resources.Reservations.CPUs)

	require.NotNil(t, source, source.Chains()[1].DockerConfig.Resources.Limits)
	require.NotNil(t, source, source.Chains()[1].DockerConfig.Resources.Reservations)

	require.NotNil(t, source.Services())
	require.Empty(t, source.Chains()[0].Config["signer-endpoint"])
}

func Test_StringConfigurationSourceFromEmptyConfig(t *testing.T) {
	cfg, err := NewStringConfigurationSource("{}", "")
	require.NoError(t, err)

	require.NotEmpty(t, cfg.Hash())
	require.Empty(t, cfg.Chains())
	require.Empty(t, cfg.FederationNodes())
	require.NotNil(t, cfg.OrchestratorOptions())
}

func Test_StringConfigurationSourceWithOverrides(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfigWithOverrides(), "http://some.ethereum.node")
	require.NoError(t, err)
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")

	require.Equal(t, "http://some.ethereum.node", source.EthereumEndpoint())
	require.Equal(t, 1*time.Minute, source.OrchestratorOptions().MaxReloadTimedDelay())
	require.Equal(t, "http://some.ethereum.node", source.Chains()[0].Config["ethereum-endpoint"])
}

func Test_StringConfigurationSourceWithSigner(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfigSigner(), "http://some.ethereum.node")
	require.NoError(t, err)
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")

	require.NotNil(t, source.Services())
	require.NotNil(t, source.Services().Signer)
	require.NotNil(t, source.Services().Signer.DockerConfig)
	require.NotNil(t, source.Services().Signer.Config)

	require.Equal(t, "http://node1-signer-service-stack:7777", source.Chains()[0].Config["signer-endpoint"])
}
