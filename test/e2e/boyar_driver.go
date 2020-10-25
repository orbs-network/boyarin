package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

const DEFAULT_VCHAIN_TIMEOUT = 90 * time.Second

type boyarDependencies struct {
	keyPair           KeyConfig
	topology          []interface{}
	genesisValidators []string
	httpPort          int
}

func readOnlyChannel(vChains ...VChainArgument) chan []VChainArgument {
	vChainsChannel := make(chan []VChainArgument, 1)
	vChainsChannel <- vChains
	close(vChainsChannel)
	return vChainsChannel
}

func SetupDynamicBoyarDependencies(t *testing.T, keyPair KeyConfig, genesisValidators []string, vChains <-chan []VChainArgument) (*config.Flags, func()) {
	deps := boyarDependencies{
		keyPair:           keyPair,
		genesisValidators: genesisValidators,
		topology: []interface{}{
			map[string]interface{}{
				"OrbsAddress": keyPair.NodeAddress,
				"Ip":          "127.0.0.1",
				"Port":        4400,
			},
		},
		httpPort: 80,
	}

	return SetupDynamicBoyarDepencenciesForNetwork(t, deps, vChains)
}

func SetupDynamicBoyarDepencenciesForNetwork(t *testing.T, deps boyarDependencies, vChains <-chan []VChainArgument) (*config.Flags, func()) {
	return SetupConfigServer(t, deps.keyPair, func(managementUrl string, vchainManagementUrl string) (*string, *string) {
		configStr := managementConfigJson(deps, []VChainArgument{}, managementUrl, vchainManagementUrl)
		managementStr := "{}"

		go func() {
			for currentChains := range vChains {
				configStr = managementConfigJson(deps, currentChains, managementUrl, vchainManagementUrl)
				managementStr = vchainManagementConfig(deps, currentChains)
				fmt.Println(configStr)
				fmt.Println(managementStr)
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

func InProcessBoyar(t *testing.T, ctx context.Context, logger log.Logger, flags *config.Flags) govnr.ShutdownWaiter {
	logger.Info("starting in-process boyar")
	waiter, err := services.Execute(ctx, flags, logger)
	require.NoError(t, err)
	return waiter
}

func TempFile(t *testing.T, keyPairJson []byte) *os.File {
	file, err := ioutil.TempFile("", "boyar")
	require.NoError(t, err)
	_, err = file.WriteString(string(keyPairJson))
	require.NoError(t, err) // temp file will not be cleaned
	return file
}

type JsonMap struct {
	value map[string]interface{}
}

func (m *JsonMap) String(name string) string {
	return m.value[name].(map[string]interface{})["Value"].(string)
}

func (m *JsonMap) Uint64(name string) uint64 {
	return uint64(m.value[name].(map[string]interface{})["Value"].(float64))
}
