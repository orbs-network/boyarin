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

const DEFAULT_VCHAIN_TIMEOUT = 60 * time.Second

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
