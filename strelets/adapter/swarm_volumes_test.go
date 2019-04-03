package adapter

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDockerSwarm_getVolumeDriverOptionsDefaults(t *testing.T) {
	orchestratorOptions := OrchestratorOptions{}
	driver, options := getVolumeDriverOptions("myVolume", orchestratorOptions, 1976)

	require.Equal(t, "local", driver)
	require.Empty(t, options)
}

func TestDockerSwarm_getVolumeDriverOptionsWithStorageOptions(t *testing.T) {
	storageOptions := make(map[string]string)
	storageOptions["artist"] = "Iggy Pop"
	storageOptions["song"] = "Passenger"

	orchestratorOptions := OrchestratorOptions{
		StorageOptions: storageOptions,
	}
	driver, options := getVolumeDriverOptions("myVolume", orchestratorOptions, 1976)

	require.Equal(t, "local", driver)
	require.NotEmpty(t, options)
	require.Equal(t, "Iggy Pop", options["artist"])
	require.Equal(t, "Passenger", options["song"])
}

func TestDockerSwarm_getVolumeDriverOptionsWithRexray(t *testing.T) {
	orchestratorOptions := OrchestratorOptions{
		StorageDriver: "rexray/ebs",
	}
	driver, options := getVolumeDriverOptions("myVolume", orchestratorOptions, 1976)

	require.Equal(t, "rexray/ebs", driver)
	require.NotEmpty(t, options)
	require.Equal(t, "1976", options["size"])
}

func TestDockerSwarm_getVolumeDriverWithLocalNFS(t *testing.T) {
	storageOptions := make(map[string]string)
	storageOptions["type"] = "nfs"

	orchestratorOptions := OrchestratorOptions{
		StorageOptions: storageOptions,
	}
	driver, options := getVolumeDriverOptions("myVolume", orchestratorOptions, 1976)

	require.Equal(t, "local", driver)
	require.NotEmpty(t, options)
	require.Equal(t, "nfs", options["type"])
	require.Equal(t, "/myVolume", options["device"])
}
