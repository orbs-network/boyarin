package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getOrchestratorOptions(t *testing.T) {
	_, err := getOrchestratorOptions("")
	require.Error(t, err)

	_, err = getOrchestratorOptions("{}")
	require.NoError(t, err)

	options, err := getOrchestratorOptions(`{"storage-driver":"amazing-custom-driver", "storage-options": {"size":"932"}}`)

	require.NoError(t, err)
	require.NotEmpty(t, "amazing-custom-driver", options.StorageDriver)
	require.Equal(t, "932", options.StorageOptions["size"])
}
