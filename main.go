package main

import (
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"os"
	"strconv"
	"strings"
)

func getPeersFromConfig(peers string, peerKeys string) *strelets.PeersMap {
	peersMap := make(strelets.PeersMap)
	keys := strings.Split(peerKeys, ",")

	for i, peer := range strings.Split(peers, ",") {
		tokens := strings.Split(peer, ":")
		port, _ := strconv.ParseInt(tokens[1], 10, 16)

		peersMap[strelets.PublicKey(keys[i])] = &strelets.Peer{tokens[0], int(port)}
	}

	return &peersMap
}

func printHelp() {
	fmt.Println("strelets provision-virtual-chain [params]")
	fmt.Println("strelets remove-virtual-chain [params]")
	fmt.Println()
	flag.Usage()
}

func getProvisionVirtualChainInput() *strelets.ProvisionVirtualChainInput {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")
	prefixPtr := flagSet.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flagSet.Int("http-port", 8080, "http port")
	gossipPortPtr := flagSet.Int("gossip-port", 4400, "gossip port")
	pathToConfig := flagSet.String("config", "", "path to node config")
	peersPtr := flagSet.String("peers", "", "list of peers ips and ports")
	peerKeys := flagSet.String("peerKeys", "", "list of peer keys")

	dockerImagePtr := flagSet.String("docker-image", "orbs", "docker image name")
	dockerTagPtr := flagSet.String("docker-tag", "export", "docker image tag")
	dockerPullPtr := flagSet.Bool("pull-docker-image", false, "pull docker image from docker registry")

	flagSet.Parse(os.Args[2:])
	vchainId := strelets.VirtualChainId(*vchainPtr)

	return &strelets.ProvisionVirtualChainInput{
		Chain:            vchainId,
		HttpPort:         *httpPortPtr,
		GossipPort:       *gossipPortPtr,
		VchainConfigPath: *pathToConfig,
		Peers:            getPeersFromConfig(*peersPtr, *peerKeys),
		DockerConfig: &strelets.DockerImageConfig{
			Prefix: *prefixPtr,
			Image:  *dockerImagePtr,
			Tag:    *dockerTagPtr,
			Pull:   *dockerPullPtr,
		},
	}
}

func getRemoveVirtualChainInput() *strelets.RemoveVirtualChainInput {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")
	prefixPtr := flagSet.String("prefix", "orbs-network", "container prefix")
	flagSet.Parse(os.Args[2:])

	vchainId := strelets.VirtualChainId(*vchainPtr)

	return &strelets.RemoveVirtualChainInput{
		Chain:        vchainId,
		DockerConfig: &strelets.DockerImageConfig{Prefix: *prefixPtr},
	}
}

func main() {
	root := "_tmp"

	if len(os.Args) < 2 {
		printHelp()
	}

	switch os.Args[1] {
	case "provision-virtual-chain":
		input := getProvisionVirtualChainInput()

		str := strelets.NewStrelets(root)
		if err := str.ProvisionVirtualChain(input); err != nil {
			panic(err)
		}
	case "remove-virtual-chain":
		input := getRemoveVirtualChainInput()

		str := strelets.NewStrelets(root)
		if err := str.RemoveVirtualChain(input); err != nil {
			panic(err)
		}
	default:
		printHelp()
	}
}
