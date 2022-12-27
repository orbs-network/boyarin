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

type boyarDependencies struct {
	keyPair           KeyConfig
	topology          []interface{}
	genesisValidators []string
	httpPort          int

	storageDriver    string
	storageMountType string

	binaryUrl    string
	binarySha256 string
}

func SetupDynamicBoyarDependencies(t *testing.T, keyPair KeyConfig, genesisValidators []string) (*config.Flags, func()) {
	deps := defaultBoyarDependencies(keyPair, genesisValidators)
	return SetupDynamicBoyarDepencenciesForNetwork(t, deps)
}

func defaultBoyarDependencies(keyPair KeyConfig, genesisValidators []string) boyarDependencies {
	return boyarDependencies{
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
}

func SetupDynamicBoyarDepencenciesForNetwork(t *testing.T, deps boyarDependencies) (*config.Flags, func()) {
	return SetupConfigServer(t, deps.keyPair, func(managementUrl string) (*string, *string) {
		configStr := managementConfigJson(deps, managementUrl)
		managementStr := "{}"

		return &configStr, &managementStr
	})
}

func SetupConfigServer(t *testing.T, keyPair KeyConfig, configBuilder func(managementUrl string) (*string, *string)) (*config.Flags, func()) {
	keyPairJson, err := json.Marshal(keyPair)
	require.NoError(t, err)
	file := TempFile(t, keyPairJson)

	cfg := &managementServiceConfig{}

	ts := serveConfig(cfg)

	managementUrl := ts.URL + "/node/management"

	managementStr, _ := configBuilder(managementUrl)

	fmt.Println(*managementStr)

	cfg.managementConfig = managementStr

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
