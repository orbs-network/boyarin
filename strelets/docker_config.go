package strelets

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type DockerImageConfig struct {
	Image  string
	Tag    string
	Pull   bool
	Prefix string
}

func (c *DockerImageConfig) FullImageName() string {
	return c.Image + ":" + c.Tag
}

func getContainerName(prefix string, vchain VirtualChainId) string {
	return fmt.Sprintf("%s-vchain-%d", prefix, vchain)
}

func getDockerNetworkOptions(httpPort int, gossipPort int) (exposedPorts map[string]interface{}, portBindings map[string][]portBinding) {
	exposedPorts = make(map[string]interface{})
	exposedPorts["8080/tcp"] = struct{}{}
	exposedPorts["4400/tcp"] = struct{}{}

	portBindings = make(map[string][]portBinding)
	portBindings["8080/tcp"] = []portBinding{{"0.0.0.0", httpPort}}
	portBindings["4400/tcp"] = []portBinding{{"0.0.0.0", gossipPort}}

	return
}

func buildJSONConfig(
	imageName string,
	httpPort int,
	gossipPort int,
	absolutePathToConfig string,
	absolutePathToLogs string,
	absoluteNetworkConfigPath string,
) ([]byte, error) {

	exposedPorts, portBindings := getDockerNetworkOptions(httpPort, gossipPort)

	configMap := make(map[string]interface{})
	configMap["Image"] = imageName
	configMap["ExposedPorts"] = exposedPorts
	configMap["CMD"] = []string{
		"/opt/orbs/orbs-node",
		"--silent",
		"--config", "/opt/orbs/config/node.json",
		"--config", "/opt/orbs/config/network.json",
		"--log", "/opt/orbs/logs/node.log",
	}

	hostConfigMap := make(map[string]interface{})
	hostConfigMap["Binds"] = []string{
		absolutePathToConfig + ":/opt/orbs/config/node.json",
		absoluteNetworkConfigPath + ":/opt/orbs/config/network.json",
		absolutePathToLogs + ":/opt/orbs/logs/",
	}
	hostConfigMap["PortBindings"] = portBindings

	configMap["HostConfig"] = hostConfigMap

	return json.Marshal(configMap)
}

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

type node struct {
	Key  string
	IP   string
	Port int
}

type portBinding struct {
	HostIp   string
	HostPort int
}

func getNetworkConfig(configDir string, peers map[PublicKey]*Peer) string {
	jsonMap := make(map[string]interface{})

	var nodes []node
	for key, peer := range peers {
		nodes = append(nodes, node{string(key), peer.ip, peer.port})
	}

	jsonMap["federation-nodes"] = nodes

	path, _ := filepath.Abs(filepath.Join(configDir, "network.json"))
	json, _ := json.Marshal(jsonMap)

	ioutil.WriteFile(path, json, 0644)

	return path
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
