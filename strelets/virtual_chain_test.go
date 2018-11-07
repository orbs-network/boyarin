package strelets

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getDockerVolumes(t *testing.T) {
	chain := &VirtualChain{
		Id: 42,
		DockerConfig: &DockerImageConfig{
			ContainerNamePrefix: "node1",
		},
	}

	volumes := chain.getContainerVolumes("/tmp")

	require.NotNil(t, volumes)
	require.EqualValues(t, "/tmp/node1-chain-42/config", volumes.configRootDir)
	require.EqualValues(t, "/tmp/node1-chain-42/logs", volumes.logsDir)
	require.EqualValues(t, "/tmp/node1-chain-42/config/keys.json", volumes.keyPairConfigFile)
	require.EqualValues(t, "/tmp/node1-chain-42/config/network.json", volumes.networkConfigFile)
}
