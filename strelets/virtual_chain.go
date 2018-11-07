package strelets

import (
	"encoding/json"
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
	return fmt.Sprintf("%s-chain-%d", v.DockerConfig.Prefix, v.Id)
}

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func copyFile(source string, destination string) error {
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destination, data, 0600)
}

type FederationNode struct {
	Key  string
	IP   string
	Port int
}

func getNetworkConfigJSON(peers *PeersMap) []byte {
	jsonMap := make(map[string]interface{})

	var nodes []FederationNode
	for key, peer := range *peers {
		nodes = append(nodes, FederationNode{string(key), peer.IP, peer.Port})
	}

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
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
