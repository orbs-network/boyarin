package p1000e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

type VChainArgument struct {
	Id       int
	Disabled bool
	BasePort int
}

const basePort = 6000

func (vc VChainArgument) GossipPort() int {
	port := basePort
	if vc.BasePort != 0 {
		port = vc.BasePort
	}

	return port + vc.Id
}

func (vc VChainArgument) HttpPort() int {
	port := basePort
	if vc.BasePort != 0 {
		port = vc.BasePort
	}
	return port - vc.Id
}

func configJson(t *testing.T, topology []interface{}, genesisValidators []string, httpPort int, vChains []VChainArgument) string {
	chains := make([]interface{}, len(vChains))
	model := map[string]interface{}{
		"network": topology,
		"orchestrator": map[string]interface{}{
			"max-reload-time-delay": "1s",
			"http-port":             httpPort,
		},
		"chains": chains,
		"services": map[string]interface{}{
			"signer": map[string]interface{}{
				"Port": 7777,
				"DockerConfig": map[string]interface{}{
					"Image": "orbsnetwork/signer",
					"Tag":   "experimental",
					"Pull":  false,
				},
			},
		},
	}
	for i, id := range vChains {
		chains[i] = VChainConfig(id, genesisValidators)
	}
	jsonStr, err := json.MarshalIndent(model, "", "    ")
	require.NoError(t, err)
	return string(jsonStr)
}

func VChainConfig(vc VChainArgument, genesisValidators []string) map[string]interface{} {
	return map[string]interface{}{
		"Id":         vc.Id,
		"HttpPort":   vc.HttpPort(),
		"GossipPort": vc.GossipPort(),
		"Disabled":   vc.Disabled,
		"DockerConfig": map[string]interface{}{
			"Image": "orbsnetwork/node",
			"Tag":   "experimental",
			"Pull":  false,
		},
		"Config": map[string]interface{}{
			"active-consensus-algo":               1, // FIXME move back to lean helix
			"genesis-validator-addresses":         genesisValidators,
			"benchmark-consensus-constant-leader": genesisValidators[0],
		},
	}
}

type KeyConfig struct {
	NodeAddress    string `json:"node-address"`
	NodePrivateKey string `json:"node-private-key,omitempty"` // Very important to omit empty value to produce a valid config
}

func serveConfig(configStr *string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("configuration requested")
		_, _ = fmt.Fprint(w, *configStr)
	}))
}

func SetupDynamicBoyarDependencies(t *testing.T, keyPair KeyConfig, vChains <-chan []VChainArgument) (*config.Flags, func()) {
	return SetupDynamicBoyarDepencenciesForNetwork(t, keyPair, []interface{}{
		map[string]interface{}{
			"address": keyPair.NodeAddress,
			"ip":      "127.0.0.1",
		},
	}, []string{
		keyPair.NodeAddress,
	}, 80, vChains)
}

func SetupDynamicBoyarDepencenciesForNetwork(t *testing.T, keyPair KeyConfig,
	topology []interface{}, genesisValidators []string, httpPort int, vChains <-chan []VChainArgument) (*config.Flags, func()) {

	keyPairJson, err := json.Marshal(keyPair)
	require.NoError(t, err)
	file := TempFile(t, keyPairJson)

	configStr := configJson(t, topology, genesisValidators, httpPort, []VChainArgument{})
	go func() {
		for currentChains := range vChains {
			configStr = configJson(t, topology, genesisValidators, httpPort, currentChains)
			fmt.Println(configStr)
		}
	}()
	ts := serveConfig(&configStr)
	flags := &config.Flags{
		Timeout:           time.Minute,
		ConfigUrl:         ts.URL,
		KeyPairConfigPath: file.Name(),
		PollingInterval:   500 * time.Millisecond,
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
	return SetupDynamicBoyarDepencenciesForNetwork(t, keyPair, topology, genesisValidators, httpPort, vChainsChannel)
}

func SetupBoyarDependencies(t *testing.T, keyPair KeyConfig, vChains ...VChainArgument) (*config.Flags, func()) {
	vChainsChannel := make(chan []VChainArgument, 1)
	vChainsChannel <- vChains
	close(vChainsChannel)
	return SetupDynamicBoyarDependencies(t, keyPair, vChainsChannel)
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

func GetVChainMetrics(t helpers.TestingT, port int, vc VChainArgument) JsonMap {
	metrics := make(map[string]interface{})
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/vchains/%d/metrics", port, vc.Id))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &metrics)
	require.NoError(t, err)
	return JsonMap{value: metrics}
}

func AssertGossipServer(t helpers.TestingT, vc VChainArgument) {
	timeout := time.Second
	port := strconv.Itoa(vc.GossipPort())
	conn, err := net.DialTimeout("tcp", "127.0.0.1:"+port, timeout)
	require.NoError(t, err, "error connecting to port %d vChainId %d", port, vc.Id)
	require.NotNil(t, conn, "nil connection to port %d vChainId %d", port, vc.Id)
	err = conn.Close()
	require.NoError(t, err, "closing connection to port %d vChainId %d", port, vc.Id)
}

func AssertServiceUp(t helpers.TestingT, ctx context.Context, serviceName string) {
	orchestrator, err := adapter.NewDockerSwarm(adapter.OrchestratorOptions{}, helpers.DefaultTestLogger())
	require.NoError(t, err)

	statuses, err := orchestrator.GetStatus(ctx, 1*time.Second)
	require.NoError(t, err)

	ok := false
	for _, status := range statuses {
		if status.Name == serviceName && status.State == "started" {
			ok = true
			return
		}
	}

	require.True(t, ok, "service should be up")
}

func AssertVchainUp(t helpers.TestingT, port int, publickKey string, vc1 VChainArgument) {
	metrics := GetVChainMetrics(t, port, vc1)
	require.Equal(t, metrics.String("Node.Address"), publickKey)
	AssertGossipServer(t, vc1)
}

func AssertVchainDown(t helpers.TestingT, port int, vc1 VChainArgument) {
	res, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/vchains/%d/metrics", port, vc1.Id))
	require.NoError(t, err)
	require.EqualValues(t, http.StatusNotFound, res.StatusCode)
}
