package adapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func storeVirtualChainConfiguration(containerName string, containerRoot string, config *AppConfig) error {
	vchainVolumes := getVirtualChainDockerContainerVolumes(containerName, containerRoot)
	vchainVolumes.createDirs()

	if err := ioutil.WriteFile(vchainVolumes.keyPairConfigFile, config.KeyPair, 0644); err != nil {
		return fmt.Errorf("could not copy key pair config: %s at %s", err, vchainVolumes.keyPairConfigFile)
	}

	if err := ioutil.WriteFile(vchainVolumes.networkConfigFile, config.Network, 0644); err != nil {
		return fmt.Errorf("could not write network config: %s at %s", err, vchainVolumes.networkConfigFile)
	}

	return nil
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

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func getVirtualChainDockerContainerVolumes(containerName string, root string) *virtualChainVolumes {
	absolutePathToConfigDir := filepath.Join(root, containerName, "config")

	return &virtualChainVolumes{
		configRootDir:     absolutePathToConfigDir,
		logsDir:           filepath.Join(root, containerName, "logs"),
		keyPairConfigFile: filepath.Join(absolutePathToConfigDir, "keys.json"),
		networkConfigFile: filepath.Join(absolutePathToConfigDir, "network.json"),
	}
}
