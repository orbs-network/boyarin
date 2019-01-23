package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/test"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

type harness struct {
	s          strelets.Strelets
	configPath string
}

func newHarness(t *testing.T, docker adapter.Orchestrator) *harness {
	removeAllDockerVolumes(t)

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

// FIXME use proper chain id
func chain(i int) *strelets.VirtualChain {
	return &strelets.VirtualChain{
		Id:           42,
		HttpPort:     8080 + i,
		GossipPort:   4400 + i,
		DockerConfig: DockerConfig(fmt.Sprintf("node%d", i)),
	}
}

func (h *harness) startChain(t *testing.T) {
	for i := 1; i <= 3; i++ {
		h.startChainInstance(t, i)
	}
}

func (h *harness) stopChain(t *testing.T) {
	h.stopChains(t, 42)
}

func (h *harness) stopChains(t *testing.T, vchainIds ...int) {
	for _, vchainId := range vchainIds {
		for i := 1; i <= 3; i++ {
			h.stopChainInstance(t, vchainId, i)
		}
	}
}

func (h *harness) stopChainInstance(t *testing.T, vchainId int, i int) {
	err := h.s.RemoveVirtualChain(context.Background(), &strelets.RemoveVirtualChainInput{
		VirtualChain: &strelets.VirtualChain{
			Id:           strelets.VirtualChainId(vchainId),
			DockerConfig: DockerConfig(fmt.Sprintf("node%d", i)),
		},
	})

	require.NoError(t, err)
}

func (h *harness) startChainInstance(t *testing.T, i int) {
	localIP := test.LocalIP()
	ctx := context.Background()

	err := h.s.ProvisionVirtualChain(ctx, &strelets.ProvisionVirtualChainInput{
		VirtualChain:      chain(i),
		Peers:             Peers(localIP),
		KeyPairConfigPath: fmt.Sprintf("%s/node%d/keys.json", h.configPath, i),
	})

	require.NoError(t, err)
}

func (h *harness) getMetricsEndpoint(port int) string {
	return "http://" + test.LocalIP() + ":" + strconv.Itoa(port) + "/metrics"
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
	return h.getMetricsForPort(8081)()
}

func (h *harness) getMetricsForEndpoint(getEndpoint func() string) func() (map[string]interface{}, error) {
	return func() (map[string]interface{}, error) {
		data, err := h.httpGet(getEndpoint())
		if err != nil {
			return nil, err
		}

		metrics := make(map[string]interface{})
		if err := json.Unmarshal(data, &metrics); err != nil {
			return nil, err
		}

		return metrics, nil
	}
}

func (h *harness) getMetricsForVchain(port int, vchainId int) func() (map[string]interface{}, error) {
	return h.getMetricsForEndpoint(func() string {
		return "http://" + test.LocalIP() + ":" + strconv.Itoa(port) + "/vchains/" + strconv.Itoa(vchainId) + "/metrics"
	})
}

func getBlockHeight(getMetrics func() (map[string]interface{}, error)) (int, error) {
	metrics, err := getMetrics()
	if err != nil {
		return 0, err
	}

	blockHeight := int(metrics["BlockStorage.BlockHeight"].(map[string]interface{})["Value"].(float64))
	fmt.Println("blockHeight", blockHeight)
	return blockHeight, nil
}

func (h *harness) getMetricsForPort(httpPort int) func() (map[string]interface{}, error) {
	return h.getMetricsForEndpoint(func() string {
		return h.getMetricsEndpoint(httpPort)
	})
}

func DockerConfig(node string) strelets.DockerImageConfig {
	return strelets.DockerImageConfig{
		Image:               "orbs",
		Tag:                 "export",
		Pull:                false,
		ContainerNamePrefix: node,
	}
}

func Peers(ip string) *strelets.PeersMap {
	peers := make(strelets.PeersMap)

	for i, key := range test.NodeAddresses() {
		peers[strelets.NodeAddress(key)] = &strelets.Peer{
			IP:   ip,
			Port: 4400 + i + 1,
		}
	}

	return &peers
}

func waitForBlock(t *testing.T, getMetrics func() (map[string]interface{}, error), targetBlockHeight int, timeout time.Duration) {
	require.True(t, test.Eventually(timeout, func() bool {
		blockHeight, err := getBlockHeight(getMetrics)
		if err != nil {
			return false
		}

		return blockHeight >= targetBlockHeight
	}))
}

func skipUnlessDockerIsEnabled(t *testing.T) {
	if os.Getenv("ENABLE_DOCKER") != "true" {
		t.Skip("skipping test, docker is disabled")
	}
}

func skipUnlessSwarmIsEnabled(t *testing.T) {
	if os.Getenv("ENABLE_SWARM") != "true" {
		t.Skip("skipping test, docker swarm is disabled")
	}
}

func removeAllDockerVolumes(t *testing.T) {
	if !(os.Getenv("ENABLE_DOCKER") == "true" || os.Getenv("ENABLE_SWARM") == "true") {
		return
	}

	t.Log("Removing all docker volumes")

	ctx := context.Background()
	client, err := client.NewClientWithOpts(client.WithVersion(adapter.DOCKER_API_VERSION))
	if err != nil {
		t.Errorf("could not connect to docker: %s", err)
		t.FailNow()
	}

	if containers, err := client.ContainerList(ctx, types.ContainerListOptions{}); err != nil {
		t.Errorf("could not list docker containers: %s", err)
		t.FailNow()
	} else {
		for _, c := range containers {
			t.Log("container", c.Names[0], "is still up with state", c.State)
		}
	}

	if volumes, err := client.VolumeList(ctx, filters.Args{}); err != nil {
		t.Errorf("could not list docker volumes: %s", err)
		t.FailNow()
	} else {
		for _, v := range volumes.Volumes {
			fmt.Println("removing volume:", v.Name)

			if err := strelets.Try(ctx, 10, 1*time.Second, 100*time.Millisecond, func(ctxWithTimeout context.Context) error {
				return client.VolumeRemove(ctx, v.Name, true)
			}); err != nil {
				t.Errorf("could not list docker volumes: %s", err)
				t.FailNow()
			}
		}
	}
}
