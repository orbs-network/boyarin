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
	"strings"
	"testing"
	"text/template"
	"time"
)

const configJsonStr = `{
  "network": [
    {
      "address":"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
      "ip":"127.0.0.1"
    }
  ],
  "orchestrator": {
    "max-reload-time-delay": "1s"
  },
  "chains": [
    {
      "Id":         {{.VChainId}},
      "HttpPort":   8080,
      "GossipPort": 4400,
      "DockerConfig": {
        "ContainerNamePrefix": "e2e",
        "Image":  "orbs",
        "Tag":    "export",
        "Pull":   false
      },
      "Config": {
        "active-consensus-algo": 2,
        "genesis-validator-addresses" : [
            "dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173"
        ]
      }
    }
  ],
  "services": {}
}`

var configJsonTemplate = template.Must(template.New("").Parse(configJsonStr))

type KeyConfig struct {
	NodeAddress    string `json:"node-address"`
	NodePrivateKey string `json:"node-private-key,omitempty"` // Very important to omit empty value to produce a valid config
}

func serveConfig(t *testing.T, vChainId int) *httptest.Server {
	var sb strings.Builder
	err := configJsonTemplate.Execute(&sb, struct {
		VChainId int
	}{
		VChainId: vChainId,
	})
	require.NoError(t, err)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//logger.Info("configuration requested")
		_, _ = fmt.Fprint(w, sb.String())
	}))
}

/**
set up environment and run boyar

*/
func InProcessBoyar(t *testing.T, logger log.Logger, keyPair KeyConfig, vChainId int) {
	keyPairJson, err := json.Marshal(keyPair)
	require.NoError(t, err)

	file, err := TempFile(err, t, keyPairJson)
	defer os.Remove(file.Name())
	ts := serveConfig(t, vChainId)
	defer ts.Close()
	flags := &config.Flags{
		Timeout:           time.Minute,
		ConfigUrl:         ts.URL,
		KeyPairConfigPath: file.Name(),
		PollingInterval:   500 * time.Millisecond,
	}
	logger.Info("starting in-process boyar")
	err = services.Execute(flags, logger)
	require.NoError(t, err)
}

func TempFile(err error, t *testing.T, keyPairJson []byte) (*os.File, error) {
	file, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = file.WriteString(string(keyPairJson))
	return file, err
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
	body, err2 := ioutil.ReadAll(resp.Body)
	require.NoError(t, err2)
	err = json.Unmarshal(body, &metrics)
	require.NoError(t, err)
	return JsonMap{value: metrics}
}
