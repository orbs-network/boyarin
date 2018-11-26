package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

type harness struct {
	s          strelets.Strelets
	configPath string
}

func newHarness(t *testing.T, docker adapter.Orchestrator) *harness {
	configPath := "../../e2e-config/"
	if configPathFromEnv := os.Getenv("E2E_CONFIG"); configPathFromEnv != "" {
		configPath = configPathFromEnv
	}

	root := "_tmp"
	os.RemoveAll(root)

	return &harness{
		s:          strelets.NewStrelets(root, docker),
		configPath: configPath,
	}
}

func chain(i int) *strelets.VirtualChain {
	return &strelets.VirtualChain{
		Id:           42,
		HttpPort:     8080 + i,
		GossipPort:   4400 + i,
		DockerConfig: DockerConfig(fmt.Sprintf("node%d", i)),
	}
}

func (h *harness) startChain(t *testing.T) {
	localIP := test.LocalIP()
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		err := h.s.ProvisionVirtualChain(ctx, &strelets.ProvisionVirtualChainInput{
			VirtualChain:      chain(i),
			Peers:             Peers(localIP),
			KeyPairConfigPath: fmt.Sprintf("%s/node%d/keys.json", h.configPath, i),
		})

		require.NoError(t, err)
	}
}

func (h *harness) stopChain(t *testing.T) {
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		h.s.RemoveVirtualChain(ctx, &strelets.RemoveVirtualChainInput{
			VirtualChain: &strelets.VirtualChain{
				Id:           42,
				DockerConfig: DockerConfig(fmt.Sprintf("node%d", i)),
			},
		})
	}
}

func (h *harness) getMetricsEndpoint() string {
	return "http://" + test.LocalIP() + ":8081/metrics"
}

func (h *harness) httpGet(url string) ([]byte, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got http status code %d calling %s", res.StatusCode, url)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (h *harness) getMetrics() (map[string]interface{}, error) {
	data, err := h.httpGet(h.getMetricsEndpoint())
	if err != nil {
		return nil, err
	}

	metrics := make(map[string]interface{})
	if err := json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}

	return metrics, nil
}

func DockerConfig(node string) *strelets.DockerImageConfig {
	return &strelets.DockerImageConfig{
		Image:               "orbs",
		Tag:                 "export",
		Pull:                false,
		ContainerNamePrefix: node,
	}
}

func Peers(ip string) *strelets.PeersMap {
	peers := make(strelets.PeersMap)

	for i, key := range test.PublicKeys() {
		peers[strelets.PublicKey(key)] = &strelets.Peer{
			IP:   ip,
			Port: 4400 + i + 1,
		}
	}

	return &peers
}

func waitForBlock(t *testing.T, getMetrics func() (map[string]interface{}, error), targetBlockHeight int, timeout time.Duration) {
	require.True(t, test.Eventually(timeout, func() bool {
		metrics, err := getMetrics()
		if err != nil {
			return false
		}

		blockHeight := int(metrics["BlockStorage.BlockHeight"].(map[string]interface{})["Value"].(float64))
		fmt.Println("blockHeight", blockHeight)

		return blockHeight >= targetBlockHeight
	}))
}
