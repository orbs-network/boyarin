package main

import (
	"flag"
	"github.com/orbs-network/boyarin/strelets"
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

func main() {
	root := "_tmp"

	vchainPtr := flag.Int("vchain", 42, "virtual chain id")
	prefixPtr := flag.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flag.Int("http-port", 8080, "http port")
	gossipPortPtr := flag.Int("gossip-port", 4400, "gossip port")
	pathToConfig := flag.String("config", "", "path to node config")
	peersPtr := flag.String("peers", "", "list of peers ips and ports")
	peerKeys := flag.String("peerKeys", "", "list of peer keys")

	dockerImagePtr := flag.String("docker-image", "orbs", "docker image name")
	dockerTagPtr := flag.String("docker-tag", "export", "docker image tag")
	dockerPullPtr := flag.Bool("pull-docker-image", false, "pull docker image from docker registry")

	flag.Parse()

	vchainId := strelets.VirtualChainId(*vchainPtr)

	str := strelets.NewStrelets(root)
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
}
