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
	configRoot    string
	keysConfig    string
	networkConfig string
	logs          string
}

func (v *VirtualChain) getDockerVolumes(root string) *virtualChainVolumes {
	containerName := v.getContainerName()
	configDir := filepath.Join(root, containerName, "config")
	absolutePathToLogs, _ := filepath.Abs(filepath.Join(root, containerName, "logs"))
	absolutePathToNetwork, _ := filepath.Abs(filepath.Join(configDir, "network.json"))
	absolutePathToKeys, _ := filepath.Abs(filepath.Join(configDir, "keys.json"))

	return &virtualChainVolumes{
		configDir,
		absolutePathToKeys,
		absolutePathToNetwork,
		absolutePathToLogs,
	}
}
