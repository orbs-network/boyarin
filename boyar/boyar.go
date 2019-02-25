package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"strings"
	"sync"
	"time"
)

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
}

type boyar struct {
	strelets          strelets.Strelets
	config            NodeConfiguration
	configCache       BoyarConfigCache
	keyPairConfigPath string
}

func NewBoyar(strelets strelets.Strelets, config NodeConfiguration, configCache BoyarConfigCache, keyPairConfigPath string) Boyar {
	return &boyar{
		strelets:          strelets,
		config:            config,
		configCache:       configCache,
		keyPairConfigPath: keyPairConfigPath,
	}
}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	var errors []error
	wg := sync.WaitGroup{}

	for _, chain := range b.config.Chains() {
		wg.Add(1)

		go func(chain *strelets.VirtualChain) {
			defer wg.Done()

			peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

			input := &strelets.ProvisionVirtualChainInput{
				VirtualChain:      chain,
				KeyPairConfigPath: b.keyPairConfigPath,
				Peers:             peers,
			}

			data, _ := json.Marshal(input)
			hash := crypto.CalculateHash(data)

			if hash == b.configCache[chain.Id.String()] {
				return
			}

			if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
				errors = append(errors, fmt.Errorf("failed to provision virtual chain %d: %s", chain.Id, err))
			} else {
				b.configCache[chain.Id.String()] = hash
				fmt.Println(time.Now(), fmt.Sprintf("INFO: updated virtual chain %d with configuration %s", chain.Id, hash))
				fmt.Println(string(data))
			}
		}(chain)
	}

	if waitTimeout(&wg, 10*time.Minute) {
		errors = append(errors, fmt.Errorf("Provisioning of virtual chains timed out"))
	}

	return aggregateErrors(errors)
}

func GetConfiguration(configUrl string, ethereumEndpoint string, topologyContractAddress string) (NodeConfiguration, error) {
	config, err := NewUrlConfigurationSource(configUrl)
	if err != nil {
		return nil, err
	}

	if ethereumEndpoint != "" && topologyContractAddress != "" {
		federationNodes, err := GetEthereumTopology(context.Background(), ethereumEndpoint, topologyContractAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to retrive topology from Ethereum: %s", err)
		}
		config.SetFederationNodes(federationNodes)
	}

	return config, err
}

func RunOnce(keyPairConfigPath string, configUrl string, ethereumEndpoint string, topologyContractAddress string, configCache BoyarConfigCache) error {
	config, err := GetConfiguration(configUrl, ethereumEndpoint, topologyContractAddress)
	if err != nil {
		return err
	}

	orchestrator, err := adapter.NewDockerSwarm(config.OrchestratorOptions())
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := NewBoyar(s, config, configCache, keyPairConfigPath)

	if err := b.ProvisionVirtualChains(context.Background()); err != nil {
		return err
	}

	if err := b.ProvisionHttpAPIEndpoint(context.Background()); err != nil {
		return err
	}

	return nil
}

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	var keys []httpReverseProxyCompositeKey

	for _, chain := range b.config.Chains() {
		keys = append(keys, httpReverseProxyCompositeKey{
			Id:         chain.Id,
			HttpPort:   chain.HttpPort,
			GossipPort: chain.HttpPort,
		})
	}

	data, _ := json.Marshal(keys)
	hash := crypto.CalculateHash(data)

	if hash == b.configCache[HTTP_REVERSE_PROXY_HASH] {
		return nil
	}

	// TODO is there a better way to get a loopback interface?
	if err := b.strelets.UpdateReverseProxy(ctx, b.config.Chains(), helpers.LocalIP()); err != nil {
		return err
	}

	b.configCache[HTTP_REVERSE_PROXY_HASH] = hash
	return nil
}

func buildPeersMap(nodes []*strelets.FederationNode, gossipPort int) *strelets.PeersMap {
	peersMap := make(strelets.PeersMap)

	for _, node := range nodes {
		// Need this override for more flexibility in network config and also for local testing
		port := node.Port
		if port == 0 {
			port = gossipPort
		}

		peersMap[strelets.NodeAddress(node.Key)] = &strelets.Peer{
			node.IP, port,
		}
	}

	return &peersMap
}

func aggregateErrors(errors []error) error {
	if errors == nil {
		return nil
	}

	var lines []string

	for _, err := range errors {
		lines = append(lines, err.Error())
	}

	return fmt.Errorf(strings.Join(lines, "\n"))
}

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
