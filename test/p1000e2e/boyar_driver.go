package p1000e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/test/helpers"
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
	Id int
}

const basePort = 6000

func (vc VChainArgument) GossipPort() int {
	return basePort + vc.Id
}

func (vc VChainArgument) HttpPort() int {
	return basePort - vc.Id
}

func configJson(t *testing.T, vChains []VChainArgument) string {
	chains := make([]interface{}, len(vChains))
	model := map[string]interface{}{
		"network": []interface{}{
			map[string]interface{}{
				"address": "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
				"ip":      "127.0.0.1",
			},
		},
		"orchestrator": map[string]interface{}{
			"max-reload-time-delay": "1s",
		},
		"chains":   chains,
		"services": map[string]interface{}{},
	}
	for i, id := range vChains {
		chains[i] = VChainConfig(id)
	}
	jsonStr, err := json.MarshalIndent(model, "", "    ")
	require.NoError(t, err)
	return string(jsonStr)
}

func VChainConfig(vc VChainArgument) map[string]interface{} {
	return map[string]interface{}{
		"Id":         vc.Id,
		"HttpPort":   vc.HttpPort(),
		"GossipPort": vc.GossipPort(),
		"DockerConfig": map[string]interface{}{
			"ContainerNamePrefix": "e2e",
			"Image":               "orbs",
			"Tag":                 "export",
			"Pull":                false,
		},
		"Config": map[string]interface{}{
			"active-consensus-algo": 2,
			"genesis-validator-addresses": []interface{}{
				"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
			},
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
	keyPairJson, err := json.Marshal(keyPair)
	require.NoError(t, err)
	file := TempFile(t, keyPairJson)
	configStr := configJson(t, []VChainArgument{})
	go func() {
		for currentChains := range vChains {
			configStr = configJson(t, currentChains)
			// fmt.Println(configStr)
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

func SetupBoyarDependencies(t *testing.T, keyPair KeyConfig, vChains ...VChainArgument) (*config.Flags, func()) {
	vChainsChannel := make(chan []VChainArgument, 1)
	vChainsChannel <- vChains
	close(vChainsChannel)
	return SetupDynamicBoyarDependencies(t, keyPair, vChainsChannel)
}

func InProcessBoyar(t *testing.T, ctx context.Context, logger log.Logger, flags *config.Flags) {
	logger.Info("starting in-process boyar")
	err := services.Execute(ctx, flags, logger)
	require.NoError(t, err)
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

func GetVChainMetrics(t helpers.TestingT, vc VChainArgument) JsonMap {
	metrics := make(map[string]interface{})
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1/vchains/%d/metrics", vc.Id))
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

func AssertVchainUp(t helpers.TestingT, publickKey string, vc1 VChainArgument) {
	metrics := GetVChainMetrics(t, vc1)
	require.Equal(t, metrics.String("Node.Address"), publickKey)
	AssertGossipServer(t, vc1)
}