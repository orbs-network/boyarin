package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeConfigurationContainer_readKeysConfig(t *testing.T) {
	source, err := parseStringConfig("{}", "")
	require.NoError(t, err)

	source.SetKeyConfigPath("./test/fake-key-pair.json")

	cfg, err := source.readKeysConfig()
	require.NoError(t, err)

	require.EqualValues(t, "cfc9e5189223aedce9543be0ef419f89aaa69e8b", cfg.Address)
	require.EqualValues(t, "c30bf9e301a19c319818b34a75901fd8f067b676a834eeb4169ec887dd03d2a8", cfg.PrivateKey)
}
