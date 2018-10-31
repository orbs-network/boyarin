package strelets

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

func createConfigDir(root string, containerName string) string {
	configDir := filepath.Join(root, containerName, "config")
	os.MkdirAll(configDir, 0755)
	return configDir
}

func createLogsDir(root string, containerName string) string {
	absoluteLogPath, _ := filepath.Abs(filepath.Join(root, containerName, "logs"))
	os.MkdirAll(absoluteLogPath, 0755)
	return absoluteLogPath
}

func copyNodeConfig(configDir string, pathToConfig string) (string, error) {
	data, err := ioutil.ReadFile(pathToConfig)

	if err != nil {
		return "", err
	}

	absolutePathToConfig, _ := filepath.Abs(filepath.Join(configDir, "config.json"))
	ioutil.WriteFile(absolutePathToConfig, data, 0600)

	return absolutePathToConfig, nil
}

type node struct {
	Key  string
	IP   string
	Port int
}

func getNetworkConfig(configDir string, peers map[PublicKey]*Peer) string {
	jsonMap := make(map[string]interface{})

	var nodes []node
	for key, peer := range peers {
		nodes = append(nodes, node{string(key), peer.IP, peer.Port})
	}

	jsonMap["federation-nodes"] = nodes

	path, _ := filepath.Abs(filepath.Join(configDir, "network.json"))
	json, _ := json.Marshal(jsonMap)

	ioutil.WriteFile(path, json, 0644)

	return path
}

type virtualChainVolumes struct {
	config string
	network string
	logs string
}

func (s *strelets) prepareVirtualChainConfig(containerName string, configPath string) *virtualChainVolumes {
	configDir := createConfigDir(s.root, containerName)
	absolutePathToLogs := createLogsDir(filepath.Join(s.root, "logs"), containerName)
	absolutePathToNetwork := getNetworkConfig(configDir, s.peers)
	absolutePathToConfig, _ := copyNodeConfig(configDir, configPath)

	return &virtualChainVolumes{
		absolutePathToConfig,
		absolutePathToNetwork,
		absolutePathToLogs,
	}
}