package p1000e2e

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/services"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/orbs-network/scribe/log"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"
)

func configJson(t *testing.T, vChainIds []int) string {
	chains := make([]interface{}, len(vChainIds))
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
	for i, id := range vChainIds {
		chains[i] = map[string]interface{}{
			"Id":         id,
			"HttpPort":   8080 + i,
			"GossipPort": 4400 + i,
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
	jsonStr, err := json.MarshalIndent(model, "", "    ")
	require.NoError(t, err)
	return string(jsonStr)
}

type KeyConfig struct {
	NodeAddress    string `json:"node-address"`
	NodePrivateKey string `json:"node-private-key,omitempty"` // Very important to omit empty value to produce a valid config
}

func serveConfig(configStr string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("configuration requested")
		_, _ = fmt.Fprint(w, configStr)
	}))
}

func SetupBoyarDependencies(t *testing.T, keyPair KeyConfig, vChainIds ...int) (*config.Flags, func()) {
	keyPairJson, err := json.Marshal(keyPair)
	require.NoError(t, err)
	file := TempFile(t, keyPairJson)
	configStr := configJson(t, vChainIds)
	// fmt.Println(configStr)
	ts := serveConfig(configStr)
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

func InProcessBoyar(t *testing.T, logger log.Logger, flags *config.Flags) {
	logger.Info("starting in-process boyar")
	err := services.Execute(flags, logger)
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

func GetVChainMetrics(t helpers.TestingT, vChainId int) JsonMap {
	metrics := make(map[string]interface{})
	resp, err := http.Get("http://127.0.0.1/vchains/" + strconv.Itoa(vChainId) + "/metrics")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	err = json.Unmarshal(body, &metrics)
	require.NoError(t, err)
	return JsonMap{value: metrics}
}
