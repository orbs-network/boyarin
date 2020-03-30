package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNodeConfigurationContainer_readKeysConfig(t *testing.T) {
	source, err := parseStringConfig("{}", "", fakeKeyPair, false)
	require.NoError(t, err)

	cfg, err := source.readKeysConfig()
	require.NoError(t, err)

	require.EqualValues(t, "cfc9e5189223aedce9543be0ef419f89aaa69e8b", cfg.Address())
	require.EqualValues(t, "c30bf9e301a19c319818b34a75901fd8f067b676a834eeb4169ec887dd03d2a8", cfg.PrivateKey())
}

func TestKeyConfig_JSON(t *testing.T) {
	keys, err := NewKeysConfig("./test/fake-key-pair.json")
	require.NoError(t, err)

	require.EqualValues(t,
		`{"node-address":"cfc9e5189223aedce9543be0ef419f89aaa69e8b","node-private-key":"c30bf9e301a19c319818b34a75901fd8f067b676a834eeb4169ec887dd03d2a8"}`,
		keys.JSON(false))

	require.EqualValues(t,
		`{"node-address":"cfc9e5189223aedce9543be0ef419f89aaa69e8b"}`,
		keys.JSON(true))
}
