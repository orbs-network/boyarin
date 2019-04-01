package boyar

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/orbs-network-go/instrumentation/log"
	"strings"
	"time"
)

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
}

type boyar struct {
	strelets    strelets.Strelets
	config      config.NodeConfiguration
	configCache config.BoyarConfigCache
	logger      log.BasicLogger
}

func NewBoyar(strelets strelets.Strelets, cfg config.NodeConfiguration, configCache config.BoyarConfigCache, logger log.BasicLogger) Boyar {
	return &boyar{
		strelets:    strelets,
		config:      cfg,
		configCache: configCache,
		logger:      logger,
	}
}

func (b *boyar) ProvisionVirtualChains(ctx context.Context) error {
	chains := b.config.Chains()

	var errors []error
	errorChannel := make(chan error, len(chains))

	for _, chain := range chains {
		if chain.Disabled {
			b.removeVirtualChain(ctx, chain, errorChannel)
		} else {
			b.provisionVirtualChain(ctx, chain, errorChannel)
		}
	}

	for i := 0; i < len(chains); i++ {
		select {
		case err := <-errorChannel:
			if err != nil {
				errors = append(errors, err)

			}
		case <-ctx.Done():
			errors = append(errors, fmt.Errorf("failed to provision virtual chain %s: %s", chains[i].Id, ctx.Err()))
		}
	}

	return aggregateErrors(errors)
}

func FullFlow(ctx context.Context, cfg config.NodeConfiguration, configCache config.BoyarConfigCache, logger log.BasicLogger) error {
	orchestrator, err := adapter.NewDockerSwarm(cfg.OrchestratorOptions())
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	s := strelets.NewStrelets(orchestrator)
	b := NewBoyar(s, cfg, configCache, logger)

	var errors []error

	if err := b.ProvisionVirtualChains(ctx); err != nil {
		errors = append(errors, err)
	}

	if err := b.ProvisionHttpAPIEndpoint(ctx); err != nil {
		errors = append(errors, err)
	}

	return aggregateErrors(errors)
}

func ReportStatus(ctx context.Context, logger log.BasicLogger) error {
	// We really don't need any options here since we're just observing
	orchestrator, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{})
	if err != nil {
		return err
	}
	defer orchestrator.Close()

	status, err := orchestrator.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to report status: %s", err)
	}

	for _, s := range status {
		if s.Error != "" {
			logger.Error("service failure",
				log.String("vchain", getVcidFromServiceName(s.Name)),
				log.String("state", s.State),
				log.Error(fmt.Errorf(s.Error)),
				log.String("logs", s.Logs))
		} else {
			logger.Info("service status",
				log.String("vchain", getVcidFromServiceName(s.Name)),
				log.String("state", s.State),
				log.String("workerId", s.NodeID),
				log.String("createdAt", formatAsISO6801(s.CreatedAt)))
		}
	}

	if len(status) == 0 {
		fmt.Println(time.Now(), "WARN: no services found")
	}

	return nil
}

func (b *boyar) ProvisionHttpAPIEndpoint(ctx context.Context) error {
	var keys []config.HttpReverseProxyCompositeKey

	// TODO move key manipulation to config package
	for _, chain := range b.config.Chains() {
		keys = append(keys, config.HttpReverseProxyCompositeKey{
			Id:         chain.Id,
			HttpPort:   chain.HttpPort,
			GossipPort: chain.HttpPort,
			Disabled:   chain.Disabled,
		})
	}

	data, _ := json.Marshal(keys)
	hash := crypto.CalculateHash(data)

	if hash == b.configCache[config.HTTP_REVERSE_PROXY_HASH] {
		return nil
	}

	// TODO is there a better way to get a loopback interface?
	if err := b.strelets.UpdateReverseProxy(ctx, b.config.Chains(), helpers.LocalIP()); err != nil {
		return err
	}

	b.configCache[config.HTTP_REVERSE_PROXY_HASH] = hash
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

		peersMap[strelets.NodeAddress(node.Address)] = &strelets.Peer{
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

func (b *boyar) removeVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan error) {
	go func() {
		input := &strelets.RemoveVirtualChainInput{
			VirtualChain: chain,
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == b.configCache[chain.Id.String()] {
			errChannel <- nil
			return
		}

		if err := b.strelets.RemoveVirtualChain(ctx, input); err != nil {
			errChannel <- fmt.Errorf("failed to remove virtual chain %d: %s", chain.Id, err)
		} else {
			b.configCache[chain.Id.String()] = hash
			b.logger.Info("removed virtual chain", log.String("chain", chain.Id.String()))
			fmt.Println(string(data))
			errChannel <- nil
		}
	}()
}

func (b *boyar) provisionVirtualChain(ctx context.Context, chain *strelets.VirtualChain, errChannel chan error) {
	go func() {
		peers := buildPeersMap(b.config.FederationNodes(), chain.GossipPort)

		input := &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain,
			KeyPairConfigPath: b.config.KeyConfigPath(),
			Peers:             peers,
			NodeAddress:       b.config.NodeAddress(),
		}

		data, _ := json.Marshal(input)
		hash := crypto.CalculateHash(data)

		if hash == b.configCache[chain.Id.String()] {
			errChannel <- nil
			return
		}

		if err := b.strelets.ProvisionVirtualChain(ctx, input); err != nil {
			errChannel <- fmt.Errorf("failed to provision virtual chain %d: %s", chain.Id, err)
		} else {
			b.configCache[chain.Id.String()] = hash
			b.logger.Info("updated virtual chain",
				log.String("chain",
					chain.Id.String()),
				log.String("configuration", string(data)))
			errChannel <- nil
		}
	}()
}

func getVcidFromServiceName(serviceName string) string {
	tokens := strings.Split(serviceName, "-")
	return tokens[len(tokens)-1]
}

func formatAsISO6801(t time.Time) string {
	return t.Format(time.RFC3339)
}
