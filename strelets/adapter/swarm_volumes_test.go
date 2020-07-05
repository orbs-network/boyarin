package adapter

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDockerSwarm_getVolumeDriverOptionsDefaults(t *testing.T) {
	orchestratorOptions := OrchestratorOptions{}
	_, options := getVolumeDriverOptions("myVolume", orchestratorOptions)

	require.Empty(t, options)
}

func TestDockerSwarm_getVolumeDriverOptionsWithStorageOptions(t *testing.T) {
	storageOptions := make(map[string]string)
	storageOptions["artist"] = "Iggy Pop"
	storageOptions["song"] = "Passenger"

	orchestratorOptions := OrchestratorOptions{
		StorageOptions: storageOptions,
	}
	_, options := getVolumeDriverOptions("myVolume", orchestratorOptions)

	require.NotEmpty(t, options)
	require.Equal(t, "Iggy Pop", options["artist"])
	require.Equal(t, "Passenger", options["song"])
}

func TestDockerSwarm_getVolumeDriverOptionsWithRexray(t *testing.T) {
	orchestratorOptions := OrchestratorOptions{
		StorageDriver: "rexray/ebs",
	}
	_, options := getVolumeDriverOptions("myVolume", orchestratorOptions)

	require.Empty(t, options) // rexray is not supported anymore
}

func TestDockerSwarm_getVolumeDriverWithLocalNFS(t *testing.T) {
	storageOptions := make(map[string]string)
	storageOptions["type"] = "nfs"

	orchestratorOptions := OrchestratorOptions{
		StorageOptions: storageOptions,
	}
	_, options := getVolumeDriverOptions("myVolume", orchestratorOptions)

	require.NotEmpty(t, options)
	require.Equal(t, "nfs", options["type"])
	require.Equal(t, "/myVolume", options["device"])
}

func TestDockerSwarm_getVolumeDriverWithBindMounts(t *testing.T) {
	storageOptions := make(map[string]string)
	storageOptions["type"] = "nfs"

	orchestratorOptions := OrchestratorOptions{
		StorageOptions:   storageOptions,
		StorageMountType: "bind",
	}
	volumeName, options := getVolumeDriverOptions("myVolume", orchestratorOptions)

	require.NotEmpty(t, options)
	require.Equal(t, "nfs", options["type"])
	require.Empty(t, options["device"])
	require.Equal(t, "/var/efs/myVolume", volumeName)
}
