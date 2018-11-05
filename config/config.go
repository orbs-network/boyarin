package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"io/ioutil"
	"strconv"
	"strings"
)

func GetProvisionVirtualChainInput(input []string) (*strelets.ProvisionVirtualChainInput, error) {
	flagSet := flag.NewFlagSet("", flag.ExitOnError)

	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")

	chainConfig := flagSet.String("chain-config", "", "path to node config in json")
	keysConfig := flagSet.String("keys-config", "", "path to public and private keys in json")
	peersConfig := flagSet.String("peers-config", "", "path to peers config in json")

	prefixPtr := flagSet.String("prefix", "orbs-network", "container prefix")
	httpPortPtr := flagSet.Int("http-port", 8080, "http port")
	gossipPortPtr := flagSet.Int("gossip-port", 4400, "gossip port")
	peersPtr := flagSet.String("peers", "", "list of peers ips and ports")
	peerKeys := flagSet.String("peerKeys", "", "list of peer keys")

	dockerImagePtr := flagSet.String("docker-image", "orbs", "docker image name")
	dockerTagPtr := flagSet.String("docker-tag", "export", "docker image tag")
	dockerPullPtr := flagSet.Bool("pull-docker-image", false, "pull docker image from docker registry")

	flagSet.Parse(input)

	var v *strelets.VirtualChain

	if *chainConfig != "" {
		jsonConfig, err := ioutil.ReadFile(*chainConfig)
		if err != nil {
			return nil, err
		}

		v = &strelets.VirtualChain{}
		if err := json.Unmarshal(jsonConfig, v); err != nil {
			return nil, err
		}
	} else {
		vchainId := strelets.VirtualChainId(*vchainPtr)

		v = &strelets.VirtualChain{
			Id:         vchainId,
			HttpPort:   *httpPortPtr,
			GossipPort: *gossipPortPtr,
			DockerConfig: &strelets.DockerImageConfig{
				Prefix: *prefixPtr,
				Image:  *dockerImagePtr,
				Tag:    *dockerTagPtr,
				Pull:   *dockerPullPtr,
			},
		}
	}

	peersMap, err := getPeersFromConfig(*peersPtr, *peerKeys, *peersConfig, v.GossipPort)
	if err != nil {
		return nil, err
	}

	return &strelets.ProvisionVirtualChainInput{
		VirtualChain:   v,
		KeysConfigPath: *keysConfig,
		Peers:          peersMap,
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

func getPeersFromConfig(peers string, peerKeys string, peersConfig string, gossipPort int) (*strelets.PeersMap, error) {
	peersMap := make(strelets.PeersMap)

	if peersConfig != "" {
		jsonConfig, err := ioutil.ReadFile(peersConfig)
		if err != nil {
			return nil, err
		}

		var nodes []strelets.FederationNode
		if err := json.Unmarshal(jsonConfig, &nodes); err != nil {
			return nil, err
		}

		peersMap = make(strelets.PeersMap)

		for _, node := range nodes {
			peersMap[strelets.PublicKey(node.Key)] = &strelets.Peer{
				node.IP, gossipPort,
			}
		}

		return &peersMap, nil
	}

	keys := strings.Split(peerKeys, ",")
	ips := strings.Split(peers, ",")

	if len(keys) != len(ips) {
		return nil, fmt.Errorf("invalid peers input, expected both keys and ips to have same length")
	}

	if peerKeys == "" {
		return &peersMap, nil
	}

	for i, peer := range ips {
		tokens := strings.Split(peer, ":")
		port, _ := strconv.ParseInt(tokens[1], 10, 16)

		peersMap[strelets.PublicKey(keys[i])] = &strelets.Peer{tokens[0], int(port)}
	}

	return &peersMap, nil
}
