package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"golang.org/x/net/context"
	"io"
	"os"
	"path/filepath"
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
	HostIp string
	HostPort string
}

func main() {
	vchainPtr := flag.String("vchain", "42", "virtual chain id")
	prefixPtr := flag.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flag.String("http-port", "8080", "http port")
	gossipPortPtr := flag.String("gossip-port", "4400", "gossip port")
	pathToConfig := flag.String("config", "", "path to node config")

	flag.Parse()

	//peersPtr := flag.String("peers", "", "list of peers ips and ports")
	//peerKeys := flag.String("peerKeys", "", "list of peer keys")

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))

	if err != nil {
		panic(err)
	}

	imageName := "orbs:export"

	//pullImage(ctx, cli, imageName)

	exposedPorts := make(map[string]interface{})
	exposedPorts[*httpPortPtr+"/tcp"] = struct {}{}
	exposedPorts[*gossipPortPtr+"/tcp"] = struct {}{}

	portBindings := make(map[string][]portBinding)
	portBindings[*httpPortPtr+"/tcp"] = []portBinding{{"0.0.0.0", "8080"}}
	portBindings[*gossipPortPtr+"/tcp"] = []portBinding{{"0.0.0.0", "4040"}}

	configMap := make(map[string]interface{})
	configMap["Image"] = imageName
	configMap["ExposedPorts"] = exposedPorts
	configMap["PortBindings"] = portBindings

	hostConfigMap := make(map[string]interface{})
	absoluteConfigPath, err := filepath.Abs(*pathToConfig)

	hostConfigMap["Binds"] = []string{absoluteConfigPath + ":/opt/orbs/config/node.json"}

	//configMap["HostConfig"] = hostConfigMap

	jsonConfig, _ := json.Marshal(configMap)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	fmt.Println(config, hostConfig, networkConfig, err)

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, getContainerName(*prefixPtr, *vchainPtr))
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)
}
