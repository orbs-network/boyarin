package config

import (
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"net/http"
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

func Test_VerifyConfigWithCorruptConfig(t *testing.T) {
	server := helpers.CreateHttpServer("/", 0, func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("{}"))
	})
	server.Start()
	defer server.Shutdown()

	source, err := GetConfiguration(&Flags{
		ConfigUrl:         server.Url(),
		KeyPairConfigPath: fakeKeyPair,
	})

	require.EqualError(t, err, "config verification failed: config is missing orchestrator options")
	require.Nil(t, source)
}
