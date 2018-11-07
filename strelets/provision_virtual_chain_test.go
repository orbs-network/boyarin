package strelets

import (
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func Test_prepareVirtualChainConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "vchain-config")
	defer os.RemoveAll(dir)

	require.NoError(t, err)

	peers := make(PeersMap)
	peers[PublicKey(test.PublicKeys()[0])] = &Peer{
		IP:   "10.0.0.1",
		Port: 4400,
	}

	input := &ProvisionVirtualChainInput{
		VirtualChain: &VirtualChain{
			Id: 42,
			DockerConfig: &DockerImageConfig{
				Prefix: "node1",
			},
		},
		Peers:             &peers,
		KeyPairConfigPath: "../e2e-config/node1/keys.json",
	}

	err = input.prepareVirtualChainConfig(dir)
	require.NoError(t, err)

	volumes := input.VirtualChain.getContainerVolumes(dir)

	require.DirExists(t, volumes.configRootDir)
	require.DirExists(t, volumes.logsDir)

	require.FileExists(t, volumes.keyPairConfigFile)
	require.FileExists(t, volumes.networkConfigFile)
}
