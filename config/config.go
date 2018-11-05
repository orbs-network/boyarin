package config

import (
	"encoding/json"
	"flag"
	"github.com/orbs-network/boyarin/strelets"
	"io/ioutil"
	"strconv"
	"strings"
)

func GetProvisionVirtualChainInput(input []string) (*strelets.ProvisionVirtualChainInput, error) {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")

	chainConfig := flagSet.String("chain-config", "", "path to node config")
	keysConfig := flagSet.String("keys-config", "", "path to public and private keys")

	prefixPtr := flagSet.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flagSet.Int("http-port", 8080, "http port")
	gossipPortPtr := flagSet.Int("gossip-port", 4400, "gossip port")
	peersPtr := flagSet.String("peers", "", "list of peers ips and ports")
	peerKeys := flagSet.String("peerKeys", "", "list of peer keys")

	dockerImagePtr := flagSet.String("docker-image", "orbs", "docker image name")
	dockerTagPtr := flagSet.String("docker-tag", "export", "docker image tag")
	dockerPullPtr := flagSet.Bool("pull-docker-image", false, "pull docker image from docker registry")

	flagSet.Parse(input)

	if *chainConfig != "" {
		jsonConfig, err := ioutil.ReadFile(*chainConfig)
		if err != nil {
			return nil, err
		}

		v := &strelets.VirtualChain{}
		if err := json.Unmarshal(jsonConfig, v); err != nil {
			return nil, err
		}

		return &strelets.ProvisionVirtualChainInput{
			VirtualChain:   v,
			KeysConfigPath: *keysConfig,
			Peers:          nil,
		}, nil
	}

	vchainId := strelets.VirtualChainId(*vchainPtr)

	return &strelets.ProvisionVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:         vchainId,
			HttpPort:   *httpPortPtr,
			GossipPort: *gossipPortPtr,
			DockerConfig: &strelets.DockerImageConfig{
				Prefix: *prefixPtr,
				Image:  *dockerImagePtr,
				Tag:    *dockerTagPtr,
				Pull:   *dockerPullPtr,
			},
		},
		KeysConfigPath: *keysConfig,
		Peers:          getPeersFromConfig(*peersPtr, *peerKeys),
	}, nil
}

func GetRemoveVirtualChainInput(input []string) *strelets.RemoveVirtualChainInput {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)
	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")
	prefixPtr := flagSet.String("prefix", "orbs-network", "container prefix")
	flagSet.Parse(input)

	vchainId := strelets.VirtualChainId(*vchainPtr)

	return &strelets.RemoveVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:           vchainId,
			DockerConfig: &strelets.DockerImageConfig{Prefix: *prefixPtr},
		},
	}
}

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
