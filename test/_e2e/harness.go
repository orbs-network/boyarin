package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func getLocalIP() net.IP {
	ifaces, _ := net.Interfaces()

	for _, i := range ifaces {
		if addrs, err := i.Addrs(); err == nil {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				if ip != nil && ip.To4() != nil && ip.To4().String() != "127.0.0.1" {
					return ip.To4()
				}
			}
		}
	}

	return nil
}

func getDockerConfigPerNode(node string) *strelets.DockerImageConfig {
	return &strelets.DockerImageConfig{
		Image:  "orbs",
		Tag:    "export",
		Pull:   false,
		Prefix: node,
	}
}

func getKeys() []string {
	return []string{
		"dfc06c5be24a67adee80b35ab4f147bb1a35c55ff85eda69f40ef827bddec173",
		"92d469d7c004cc0b24a192d9457836bf38effa27536627ef60718b00b0f33152",
		"a899b318e65915aa2de02841eeb72fe51fddad96014b73800ca788a547f8cce0",
	}
}

func getPeers(ip net.IP) *strelets.PeersMap {
	peers := make(strelets.PeersMap)

	for i, key := range getKeys() {
		peers[strelets.PublicKey(key)] = &strelets.Peer{
			IP:   ip.String(),
			Port: 4400 + i,
		}
	}

	return &peers
}

type harness struct {
	s          strelets.Strelets
	configPath string
}

func newHarness() *harness {
	configPath := "../../e2e-config/"
	if configPathFromEnv := os.Getenv("E2E_CONFIG"); configPathFromEnv != "" {
		configPath = configPathFromEnv
	}

	root := "_tmp"
	os.RemoveAll(root)

	return &harness{
		s:          strelets.NewStrelets(root),
		configPath: configPath,
	}
}

func (h *harness) startChain(t *testing.T) {
	localIP := getLocalIP()
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		err := h.s.ProvisionVirtualChain(ctx, &strelets.ProvisionVirtualChainInput{
			VirtualChain: &strelets.VirtualChain{
				Id:           42,
				HttpPort:     8080 + i,
				GossipPort:   4400 + i,
				DockerConfig: getDockerConfigPerNode(fmt.Sprintf("node%d", i+1)),
			},
			Peers:          getPeers(localIP),
			KeysConfigPath: fmt.Sprintf("%s/node%d/keys.json", h.configPath, i+1),
		})

		require.NoError(t, err)
	}
}

func (h *harness) stopChain(t *testing.T) {
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		h.s.RemoveVirtualChain(ctx, &strelets.RemoveVirtualChainInput{
			VirtualChain: &strelets.VirtualChain{
				Id:           42,
				DockerConfig: getDockerConfigPerNode(fmt.Sprintf("node%d", i+1)),
			},
		})
	}

	time.Sleep(1 * time.Second)
}

func (h *harness) getMetricsEndpoint() string {
	return "http://" + getLocalIP().String() + ":8081/metrics"
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

const eventuallyIterations = 25

func Eventually(timeout time.Duration, f func() bool) bool {
	for i := 0; i < eventuallyIterations; i++ {
		if f() {
			return true
		}
		time.Sleep(timeout / eventuallyIterations)
	}
	return false
}
