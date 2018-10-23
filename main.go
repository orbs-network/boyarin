package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func pullImage(ctx context.Context, cli *client.Client, imageName string) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
}

func getContainerName(prefix string, vchain string) string {
	return prefix + "-vchain-" + vchain
}

type portBinding struct {
	HostIp   string
	HostPort string
}

type node struct {
	Key  string
	IP   string
	Port int64
}

func getNetworkConfig(configDir string, peers string, peerKeys string) string {
	jsonMap := make(map[string]interface{})

	keys := strings.Split(peerKeys, ",")

	var nodes []node
	for i, peer := range strings.Split(peers, ",") {
		tokens := strings.Split(peer, ":")
		port, _ := strconv.ParseInt(tokens[1], 10, 16)

		nodes = append(nodes, node{keys[i], tokens[0], port})
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

func getDockerNetworkOptions(httpPort string, gossipPort string) (exposedPorts map[string]interface{}, portBindings map[string][]portBinding){
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
	httpPort string,
	gossipPort string,
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

func main() {
	root := "_tmp"

	vchainPtr := flag.String("vchain", "42", "virtual chain id")
	prefixPtr := flag.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flag.String("http-port", "8080", "http port")
	gossipPortPtr := flag.String("gossip-port", "4400", "gossip port")
	pathToConfig := flag.String("config", "", "path to node config")
	pathToLogsPtr := flag.String("logs", root, "path to logs for the virtual chain")
	peersPtr := flag.String("peers", "", "list of peers ips and ports")
	peerKeys := flag.String("peerKeys", "", "list of peer keys")

	dockerImagePtr := flag.String("docker-image", "orbs", "docker image name")
	dockerTagPtr := flag.String("docker-tag", "export", "docker image tag")
	dockerPullPtr := flag.Bool("pull-docker-image", false, "pull docker image from docker registry")

	flag.Parse()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))

	if err != nil {
		panic(err)
	}

	imageName := *dockerImagePtr + ":" + *dockerTagPtr

	if *dockerPullPtr {
		pullImage(ctx, cli, imageName)
	}

	containerName := getContainerName(*prefixPtr, *vchainPtr)
	configDir := createConfigDir(root, containerName)

	absolutePathToLogs := createLogsDir(*pathToLogsPtr, containerName)
	absoluteNetworkConfigPath := getNetworkConfig(configDir, *peersPtr, *peerKeys)
	absolutePathToConfig, err := copyNodeConfig(configDir, *pathToConfig)
	if err != nil {
		panic(err)
	}

	jsonConfig, _ := buildJSONConfig(imageName, *httpPortPtr,
		*gossipPortPtr, absolutePathToConfig, absolutePathToLogs, absoluteNetworkConfigPath)

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

	fmt.Println(resp.ID)
}
