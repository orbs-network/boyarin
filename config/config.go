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
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)

	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")

	chainConfig := flagSet.String("chain-config", "", "path to node config in json")
	keysConfig := flagSet.String("keys-config", "", "path to public and private keys in json")
	peersConfig := flagSet.String("peers-config", "", "path to peers config in json")

	prefixPtr := flagSet.String("container-name-prefix", "orbs-network", "")
	httpPortPtr := flagSet.Int("http-port", 8080, "http port")
	gossipPortPtr := flagSet.Int("gossip-port", 4400, "gossip port")
	peersPtr := flagSet.String("peers", "", "list of peers ips and ports")
	peerKeys := flagSet.String("peerKeys", "", "list of peer keys")

	dockerImagePtr := flagSet.String("docker-image", "orbs", "docker image name")
	dockerTagPtr := flagSet.String("docker-tag", "export", "docker image tag")
	dockerPullPtr := flagSet.Bool("pull-docker-image", false, "pull docker image from docker registry")

	if err := flagSet.Parse(input); err != nil {
		flag.Usage()
		return nil, err
	}

	var v *strelets.VirtualChain

	if *chainConfig != "" {
		jsonConfig, err := ioutil.ReadFile(*chainConfig)
		if err != nil {
			return nil, fmt.Errorf("virtual chain config not found: %s at %s", err, *chainConfig)
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
				ContainerNamePrefix: *prefixPtr,
				Image:               *dockerImagePtr,
				Tag:                 *dockerTagPtr,
				Pull:                *dockerPullPtr,
			},
		}
	}

	peersMap, err := getPeersFromConfig(*peersPtr, *peerKeys, *peersConfig, v.GossipPort)
	if err != nil {
		return nil, err
	}

	return &strelets.ProvisionVirtualChainInput{
		VirtualChain:      v,
		KeyPairConfigPath: *keysConfig,
		Peers:             peersMap,
	}, nil
}

func GetRemoveVirtualChainInput(input []string) (*strelets.RemoveVirtualChainInput, error) {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	vchainPtr := flagSet.Int("chain", 42, "virtual chain id")
	prefixPtr := flagSet.String("ContainerNamePrefix", "orbs-network", "container prefix")

	if err := flagSet.Parse(input); err != nil {
		flag.Usage()
		return nil, err
	}
	vchainId := strelets.VirtualChainId(*vchainPtr)

	return &strelets.RemoveVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:           vchainId,
			DockerConfig: &strelets.DockerImageConfig{ContainerNamePrefix: *prefixPtr},
		},
	}, nil
}

func getPeersFromConfig(peers string, peerKeys string, peersConfig string, gossipPort int) (*strelets.PeersMap, error) {
	peersMap := make(strelets.PeersMap)

	if peersConfig != "" {
		jsonConfig, err := ioutil.ReadFile(peersConfig)
		if err != nil {
			return nil, fmt.Errorf("virtual chain network config not found: %s at %s", err, peersConfig)
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
