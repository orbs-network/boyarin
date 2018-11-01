package strelets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type vchain struct {
	id           VirtualChainId
	httpPort     int
	gossipPort   int
	dockerConfig *DockerImageConfig
}

func (v *vchain) getContainerName() string {
	return fmt.Sprintf("%s-vchain-%d", v.dockerConfig.Prefix, v.id)
}

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func copyNodeConfig(source string, destination string) error {
	data, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(destination, data, 0600)
}

type node struct {
	Key  string
	IP   string
	Port int
}

func getNetworkConfigJSON(peers *PeersMap) []byte {
	jsonMap := make(map[string]interface{})

	var nodes []node
	for key, peer := range *peers {
		nodes = append(nodes, node{string(key), peer.IP, peer.Port})
	}

	jsonMap["federation-nodes"] = nodes
	json, _ := json.Marshal(jsonMap)

	return json
}

type virtualChainVolumes struct {
	configRoot string
	config     string
	network    string
	logs       string
}

func (s *strelets) prepareVirtualChainConfig(containerName string) *virtualChainVolumes {
	configDir := filepath.Join(s.root, containerName, "config")
	absolutePathToLogs, _ := filepath.Abs(filepath.Join(s.root, containerName, "logs"))
	absolutePathToNetwork, _ := filepath.Abs(filepath.Join(configDir, "network.json"))
	absolutePathToConfig, _ := filepath.Abs(filepath.Join(configDir, "config.json"))

	return &virtualChainVolumes{
		configDir,
		absolutePathToConfig,
		absolutePathToNetwork,
		absolutePathToLogs,
	}
}
