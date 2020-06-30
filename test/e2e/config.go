package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var NETWORK_KEY_CONFIG = []KeyConfig{
	{
		"a328846cd5b4979d68a8c58a9bdfeee657b34de7",
		"901a1a0bfbe217593062a054e561e708707cb814a123474c25fd567a0fe088f8",
	},
	{
		"d27e2e7398e2582f63d0800330010b3e58952ff6",
		"87a210586f57890ae3642c62ceb58f0f0a54e787891054a5a54c80e1da418253",
	},
	{
		"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9",
		"426308c4d11a6348a62b4fdfb30e2cad70ab039174e2e8ea707895e4c644c4ec",
	},
	{
		"c056dfc0d1fbc7479db11e61d1b0b57612bf7f17",
		"1e404ba4e421cedf58dcc3dddcee656569afc7904e209612f7de93e1ad710300",
	},
}

func genesisValidators(keyPairs []KeyConfig) (genesisValidators []string) {
	for _, keyPair := range keyPairs {
		genesisValidators = append(genesisValidators, keyPair.NodeAddress)
	}

	return
}

type KeyConfig struct {
	NodeAddress    string `json:"node-address"`
	NodePrivateKey string `json:"node-private-key,omitempty"` // Very important to omit empty value to produce a valid config
}

type managementServiceConfig struct {
	managementConfig       *string
	vchainManagementConfig *string
}

func serveConfig(serviceConfig *managementServiceConfig) *httptest.Server {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", helpers.LocalIP()))
	if err != nil {
		panic(err)
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/node/management", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, *serviceConfig.managementConfig)
	})

	handler.HandleFunc("/vchains/any/management", func(w http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(w, *serviceConfig.vchainManagementConfig)
	})

	server := &httptest.Server{
		Listener: l,
		Config: &http.Server{
			Handler: handler,
		},
	}
	server.Start()

	return server
}

func SetupDynamicBoyarDependencies(t *testing.T, keyPair KeyConfig, genesisValidators []string, vChains <-chan []VChainArgument) (*config.Flags, func()) {
	return SetupDynamicBoyarDepencenciesForNetwork(t, keyPair, genesisValidators, []interface{}{
		map[string]interface{}{
			"OrbsAddress": keyPair.NodeAddress,
			"Ip":          "127.0.0.1",
		},
	}, 80, vChains)
}

func SetupDynamicBoyarDepencenciesForNetwork(t *testing.T, keyPair KeyConfig, genesisValidators []string,
	topology []interface{}, httpPort int, vChains <-chan []VChainArgument) (*config.Flags, func()) {
	return SetupConfigServer(t, keyPair, func(managementUrl string, vchainManagementUrl string) (*string, *string) {
		configStr := managementConfigJson(managementUrl, vchainManagementUrl, httpPort, []VChainArgument{})
		managementStr := "{}"

		go func() {
			for currentChains := range vChains {
				configStr = managementConfigJson(managementUrl, vchainManagementUrl, httpPort, currentChains)
				managementStr = vchainManagementConfig(currentChains, topology, genesisValidators)
				fmt.Println(configStr)
			}
		}()

		return &configStr, &managementStr
	})
}

func SetupConfigServer(t *testing.T, keyPair KeyConfig, configBuilder func(managementUrl string, vchainManagementUrl string) (*string, *string)) (*config.Flags, func()) {
	keyPairJson, err := json.Marshal(keyPair)
	require.NoError(t, err)
	file := TempFile(t, keyPairJson)

	cfg := &managementServiceConfig{}

	ts := serveConfig(cfg)

	managementUrl := ts.URL + "/node/management"
	vchainsManagentUrl := ts.URL + "/vchains/any/management"

	managementStr, vchainsManagementStr := configBuilder(managementUrl, vchainsManagentUrl)
	fmt.Println(*managementStr, *vchainsManagementStr)

	cfg.managementConfig = managementStr
	cfg.vchainManagementConfig = vchainsManagementStr

	flags := &config.Flags{
		Timeout:           time.Minute,
		ConfigUrl:         managementUrl,
		KeyPairConfigPath: file.Name(),
		PollingInterval:   500 * time.Millisecond,
		WithNamespace:     true,
	}
	cleanup := func() {
		defer os.Remove(file.Name())
		defer ts.Close()
	}
	return flags, cleanup
}

func SetupBoyarDependenciesForNetwork(t *testing.T, keyPair KeyConfig, topology []interface{}, genesisValidators []string, httpPort int, vChains ...VChainArgument) (*config.Flags, func()) {
	vChainsChannel := make(chan []VChainArgument, 1)
	vChainsChannel <- vChains
	close(vChainsChannel)
	return SetupDynamicBoyarDepencenciesForNetwork(t, keyPair, genesisValidators, topology, httpPort, vChainsChannel)
}

func SetupBoyarDependencies(t *testing.T, keyPair KeyConfig, genesisValidators []string, vChains ...VChainArgument) (*config.Flags, func()) {
	vChainsChannel := make(chan []VChainArgument, 1)
	vChainsChannel <- vChains
	close(vChainsChannel)
	return SetupDynamicBoyarDependencies(t, keyPair, genesisValidators, vChainsChannel)
}

func InProcessBoyar(t *testing.T, ctx context.Context, logger log.Logger, flags *config.Flags) govnr.ShutdownWaiter {
	logger.Info("starting in-process boyar")
	waiter, err := services.Execute(ctx, flags, logger)
	require.NoError(t, err)
	return waiter
}

// Every vc is on the same shared network -
// a quirk of the shared networks that WILL stop working if shared network name becomes unique
func buildTopology(keyPairs []KeyConfig, vcId int) (topology []interface{}) {
	for _, kp := range keyPairs {
		topology = append(topology, map[string]interface{}{
			"OrbsAddress": kp.NodeAddress,
			"Ip":          fmt.Sprintf("%s-chain-%d", config.NodeAddress(kp.NodeAddress).ShortID(), vcId),
			"Port":        4400,
		})
	}

	return
}
