package main

import (
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"os"
	"strconv"
	"strings"
)

func getPeersFromConfig(peers string, peerKeys string) map[strelets.PublicKey]*strelets.Peer {
	peersMap := make(map[strelets.PublicKey]*strelets.Peer)
	keys := strings.Split(peerKeys, ",")

	for i, peer := range strings.Split(peers, ",") {
		tokens := strings.Split(peer, ":")
		port, _ := strconv.ParseInt(tokens[1], 10, 16)

		peersMap[strelets.PublicKey(keys[i])] = &strelets.Peer{tokens[0], int(port)}
	}

	return peersMap
}

func printHelp() {
	fmt.Println("strelets provision-virtual-chain [params]")
	fmt.Println("strelets remove-virtual-chain [params]")
	fmt.Println()
	flag.Usage()
}

func main() {
	root := "_tmp"

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

	if len(os.Args) < 2 {
		printHelp()
	}

	flagSet.Parse(os.Args[2:])

	vchainId := strelets.VirtualChainId(*vchainPtr)

	switch os.Args[1] {
	case "provision-virtual-chain":
		str := strelets.NewStrelets(root)

		fmt.Println("ZZZ", *peersPtr, *peerKeys)

		str.UpdateFederation(getPeersFromConfig(*peersPtr, *peerKeys))

		err := str.ProvisionVirtualChain(vchainId, *pathToConfig, *httpPortPtr, *gossipPortPtr,
			&strelets.DockerImageConfig{
				Image:  *dockerImagePtr,
				Tag:    *dockerTagPtr,
				Pull:   *dockerPullPtr,
				Prefix: *prefixPtr,
			},
		)

		if err != nil {
			panic(err)
		}
	default:
		printHelp()
	}
}
