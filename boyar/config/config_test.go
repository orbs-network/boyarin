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

	require.NotNil(t, source.OrchestratorOptions())
	require.Equal(t, "ebs", source.OrchestratorOptions().StorageDriver)
	require.NotNil(t, source.OrchestratorOptions().StorageOptions)
	require.Equal(t, "100", source.OrchestratorOptions().StorageOptions["maxRetries"])

	require.NotNil(t, source.Services())

}

func Test_StringConfigurationSourceFromEmptyConfig(t *testing.T) {
	cfg, err := NewStringConfigurationSource(`{"orchestrator":{}}`, "", fakeKeyPair, false)
	require.NoError(t, err)

	require.NotEmpty(t, cfg.Hash())
	require.Empty(t, cfg.FederationNodes())
	require.NotNil(t, cfg.OrchestratorOptions())
}

func Test_StringConfigurationSourceWithOverrides(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, false)
	require.NoError(t, err)

	require.Equal(t, "http://some.ethereum.node", source.EthereumEndpoint())
	require.Equal(t, 1*time.Minute, source.OrchestratorOptions().MaxReloadTimedDelay())
}

func Test_StringConfigurationSourceWithSigner(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, false)
	require.NoError(t, err)

	require.NotNil(t, source.Services())
	require.NotNil(t, source.Services().Signer())
	require.NotNil(t, source.Services().Signer().DockerConfig)
	require.NotNil(t, source.Services().Signer().Config)

}

func Test_StringConfigurationSourceWithSignerWithNamespace(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig(), "http://some.ethereum.node", fakeKeyPair, true)
	require.NoError(t, err)

	require.NotNil(t, source.Services())
	require.NotNil(t, source.Services().Signer())
	require.NotNil(t, source.Services().Signer().DockerConfig)
	require.NotNil(t, source.Services().Signer().Config)

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
	require.Empty(t, signer.ExecutablePath)

	customService := source.Services()["service-name"]
	require.NotNil(t, customService)
	require.NotNil(t, customService.DockerConfig)
	require.NotNil(t, customService.Config)

	require.False(t, customService.AllowAccessToSigner)
	require.True(t, customService.AllowAccessToServices)
	require.Empty(t, customService.ExecutablePath)
}
