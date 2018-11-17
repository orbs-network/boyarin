package adapter

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func (d *dockerAPI) StoreConfiguration(ctx context.Context, containerName string, containerRoot string, config *AppConfig) (interface{}, error) {
	return nil, storeConfiguration(containerName, containerRoot, config)
}

func storeConfiguration(containerName string, containerRoot string, config *AppConfig) error {
	vchainVolumes := getDockerContainerVolumes(containerName, containerRoot)
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

func getDockerContainerVolumes(containerName string, root string) *virtualChainVolumes {
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
