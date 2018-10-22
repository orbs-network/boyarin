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

func getNetworkConfig(containerName string, peers string, peerKeys string) string {
	jsonMap := make(map[string]interface{})

	keys := strings.Split(peerKeys, ",")

	var nodes []node
	for i, peer := range strings.Split(peers, ",") {
		tokens := strings.Split(peer, ":")
		port, _ := strconv.ParseInt(tokens[1], 10, 16)

		nodes = append(nodes, node{keys[i], tokens[0], port})
	}

	jsonMap["federation-nodes"] = nodes

	os.MkdirAll("_tmp", 0755)

	path, _ := filepath.Abs("_tmp/" + containerName + ".json")
	json, _ := json.Marshal(jsonMap)

	ioutil.WriteFile(path, json, 0644)

	return path
}

func buildJSONConfig(containerName string, imageName string, vchain string, httpPort string, gossipPort string, pathToConfig string, peers string, peerKeys string) ([]byte, error) {
	exposedPorts := make(map[string]interface{})
	exposedPorts["8080/tcp"] = struct{}{}
	exposedPorts["4400/tcp"] = struct{}{}

	portBindings := make(map[string][]portBinding)
	portBindings["8080/tcp"] = []portBinding{{"0.0.0.0", httpPort}}
	portBindings["4400/tcp"] = []portBinding{{"0.0.0.0", gossipPort}}

	configMap := make(map[string]interface{})
	configMap["Image"] = imageName
	configMap["ExposedPorts"] = exposedPorts
	configMap["CMD"] = []string{
		"/opt/orbs/orbs-node",
		"--config", "/opt/orbs/config/node.json",
		"--config", "/opt/orbs/config/network.json",
		//"--log", "/opt/orbs/logs/node.log",
	}

	hostConfigMap := make(map[string]interface{})
	configMap["HostConfig"] = hostConfigMap

	absoluteConfigPath, err := filepath.Abs(pathToConfig)
	if err != nil {
		return nil, err
	}

	absoluteNetworkConfigPath := getNetworkConfig(containerName, peers, peerKeys)
	hostConfigMap["Binds"] = []string{
		absoluteConfigPath + ":/opt/orbs/config/node.json",
		absoluteNetworkConfigPath + ":/opt/orbs/config/network.json",
	}
	hostConfigMap["PortBindings"] = portBindings

	return json.Marshal(configMap)
}

func main() {
	vchainPtr := flag.String("vchain", "42", "virtual chain id")
	prefixPtr := flag.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flag.String("http-port", "8080", "http port")
	gossipPortPtr := flag.String("gossip-port", "4400", "gossip port")
	pathToConfig := flag.String("config", "", "path to node config")
	peersPtr := flag.String("peers", "", "list of peers ips and ports")
	peerKeys := flag.String("peerKeys", "", "list of peer keys")

	flag.Parse()

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))

	if err != nil {
		panic(err)
	}

	imageName := "orbs:export"

	//pullImage(ctx, cli, imageName)

	containerName := getContainerName(*prefixPtr, *vchainPtr)

	jsonConfig, _ := buildJSONConfig(containerName, imageName, *vchainPtr, *httpPortPtr, *gossipPortPtr, *pathToConfig, *peersPtr, *peerKeys)
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
