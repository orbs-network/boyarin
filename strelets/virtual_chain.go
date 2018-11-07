package strelets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type VirtualChainId uint32

type VirtualChain struct {
	Id           VirtualChainId
	HttpPort     int
	GossipPort   int
	DockerConfig *DockerImageConfig
}

func (v *VirtualChain) getContainerName() string {
	return fmt.Sprintf("%s-chain-%d", v.DockerConfig.ContainerNamePrefix, v.Id)
}

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func copyFile(source string, destination string) error {
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("%s: %s", err, source)
	}

	return ioutil.WriteFile(destination, data, 0600)
}

type virtualChainVolumes struct {
	configRootDir string
	logsDir       string

	keyPairConfigFile string
	networkConfigFile string
}

func (v *virtualChainVolumes) createDirs() {
	createDir(v.configRootDir)
	createDir(v.logsDir)
}

func (v *VirtualChain) getContainerVolumes(root string) *virtualChainVolumes {
	containerName := v.getContainerName()

	absolutePathToConfigDir := filepath.Join(root, containerName, "config")
	absolutePathToLogDir, _ := filepath.Abs(filepath.Join(root, containerName, "logs"))

	absolutePathToNetworkConfig, _ := filepath.Abs(filepath.Join(absolutePathToConfigDir, "network.json"))
	absolutePathToKeyPairConfig, _ := filepath.Abs(filepath.Join(absolutePathToConfigDir, "keys.json"))

	return &virtualChainVolumes{
		configRootDir:     absolutePathToConfigDir,
		logsDir:           absolutePathToLogDir,
		keyPairConfigFile: absolutePathToKeyPairConfig,
		networkConfigFile: absolutePathToNetworkConfig,
	}
}
