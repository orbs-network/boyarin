package adapter

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func Test_prepareVirtualChainConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "vchain-config")
	defer os.RemoveAll(dir)

	require.NoError(t, err)

	err = storeConfiguration("fake-container-name", dir, &AppConfig{
		KeyPair: []byte("some-keys"),
		Network: []byte("some-network-config"),
	})
	require.NoError(t, err)

	volumes := getContainerVolumes("fake-container-name", dir)

	require.DirExists(t, volumes.configRootDir)
	require.DirExists(t, volumes.logsDir)

	require.FileExists(t, volumes.keyPairConfigFile)
	require.FileExists(t, volumes.networkConfigFile)
}
