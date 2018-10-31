package strelets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type VirtualChainId uint32

type PublicKey string

type Strelets interface {
	GetChain(chain VirtualChainId) (*vchain, error)
	ProvisionVirtualChain(chain VirtualChainId, configPath string, httpPort int, gossipPort int, dockerConfig *DockerImageConfig) error
	UpdateFederation(peers map[PublicKey]*Peer)
}

type strelets struct {
	vchains map[VirtualChainId]*vchain
	peers   map[PublicKey]*Peer

	root string
}

type vchain struct {
	id           VirtualChainId
	httpPort     int
	gossipPort   int
	dockerConfig *DockerImageConfig
}

type Peer struct {
	ip   string
	port int
}

type DockerImageConfig struct {
	Image  string
	Tag    string
	Pull   bool
	Prefix string
}

func NewStrelets(root string) Strelets {
	return &strelets{
		vchains: make(map[VirtualChainId]*vchain),
		peers:   make(map[PublicKey]*Peer),
		root:    root,
	}
}

func (s *strelets) GetChain(chain VirtualChainId) (*vchain, error) {
	if v, found := s.vchains[chain]; !found {
		return v, errors.Errorf("virtual chain with id %h not found", chain)
	} else {
		return v, nil
	}
}

func (s *strelets) UpdateFederation(peers map[PublicKey]*Peer) {
	s.peers = peers
}

func (s *strelets) ProvisionVirtualChain(chain VirtualChainId, configPath string, httpPort int, gossipPort int, dockerConfig *DockerImageConfig) error {
	v := &vchain{
		id:           chain,
		httpPort:     httpPort,
		gossipPort:   gossipPort,
		dockerConfig: dockerConfig,
	}
	s.vchains[chain] = v

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))

	if err != nil {
		panic(err)
	}

	imageName := v.dockerConfig.Image + ":" + v.dockerConfig.Tag

	if v.dockerConfig.Pull {
		pullImage(ctx, cli, imageName)
	}

	containerName := getContainerName(v.dockerConfig.Prefix, v.id)
	configDir := createConfigDir(s.root, containerName)

	absolutePathToLogs := createLogsDir(filepath.Join(s.root, "logs"), containerName)
	absoluteNetworkConfigPath := getNetworkConfig(configDir, s.peers)
	absolutePathToConfig, err := copyNodeConfig(configDir, configPath)
	if err != nil {
		panic(err)
	}

	jsonConfig, _ := buildJSONConfig(imageName, v.httpPort, v.gossipPort,
		absolutePathToConfig, absolutePathToLogs, absoluteNetworkConfigPath)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, containerName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return nil
}

func pullImage(ctx context.Context, cli *client.Client, imageName string) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
}

func getContainerName(prefix string, vchain VirtualChainId) string {
	return fmt.Sprintf("%s-vchain-%d", prefix, vchain)
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
