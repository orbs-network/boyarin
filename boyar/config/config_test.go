package config

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func getTestJSONConfig() string {
	contents, err := ioutil.ReadFile("./test/config.json")
	if err != nil {
		panic(err)
	}

	return string(contents)
}

func Test_StringConfigurationSource(t *testing.T) {
	source, err := NewStringConfigurationSource(getTestJSONConfig())
	source.SetKeyConfigPath("/tmp/fake-key-pair.json")
	require.NoError(t, err)

	require.NotEmpty(t, source.Hash())

	require.Equal(t, "http://localhost:8545", source.Chains()[0].Config["ethereum-endpoint"])

	require.NotNil(t, source.OrchestratorOptions())
	require.Equal(t, "ebs", source.OrchestratorOptions().StorageDriver)
	require.NotNil(t, source.OrchestratorOptions().StorageOptions)
	require.Equal(t, "100", source.OrchestratorOptions().StorageOptions["maxRetries"])
}

func Test_StringConfigurationSourceFromEmptyConfig(t *testing.T) {
	cfg, err := NewStringConfigurationSource("{}")
	require.NoError(t, err)

	require.NotEmpty(t, cfg.Hash())
	require.Empty(t, cfg.Chains())
	require.Empty(t, cfg.FederationNodes())
	require.NotNil(t, cfg.OrchestratorOptions())
}
