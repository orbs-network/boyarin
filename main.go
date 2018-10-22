package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"github.com/docker/go-connections/nat"
	"github.com/docker/docker/runconfig"
	"io"
	"os"
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

func main() {
	vchainPtr := flag.String("vchain", "42", "virtual chain id")
	prefixPtr := flag.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flag.String("http-port", "8080", "http port")
	gossipPortPtr := flag.String("gossip-port", "4400", "gossip port")
	//pathToConfig := flag.String("pathToConfig", "", "path to node config")

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

	portSet, _, _ := nat.ParsePortSpecs([]string{*httpPortPtr+":8080", *gossipPortPtr+":4400"})



	//decoder := runconfig.ContainerDecoder{}
	//config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(b))

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: portSet,
	}, nil, nil, getContainerName(*prefixPtr, *vchainPtr))
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)
}
