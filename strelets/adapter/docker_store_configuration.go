package adapter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

	if err := ioutil.WriteFile(vchainVolumes.generalConfigFile, config.Config, 0644); err != nil {
		return fmt.Errorf("count not write general config: %s at %s", err, vchainVolumes.generalConfigFile)
	}

	return nil
}

type virtualChainVolumes struct {
	configRootDir string

	keyPairConfigFile string
	networkConfigFile string
	generalConfigFile string
}

func (v *virtualChainVolumes) createDirs() {
	createDir(v.configRootDir)
}

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func getVirtualChainDockerContainerVolumes(containerName string, root string) *virtualChainVolumes {
	absolutePathToConfigDir := filepath.Join(root, containerName, "config")

	return &virtualChainVolumes{
		configRootDir:     absolutePathToConfigDir,
		keyPairConfigFile: filepath.Join(absolutePathToConfigDir, "keys.json"),
		networkConfigFile: filepath.Join(absolutePathToConfigDir, "network.json"),
		generalConfigFile: filepath.Join(absolutePathToConfigDir, "config.json"),
	}
}

func storeNginxConfiguration(nginxConfigDir string, config string) error {
	os.MkdirAll(nginxConfigDir, 0755)

	if err := ioutil.WriteFile(path.Join(nginxConfigDir, "nginx.conf"), []byte(config), 0644); err != nil {
		return fmt.Errorf("could not save nginx configuration: %s", err)
	}

	return nil
}
