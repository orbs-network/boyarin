package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/test/helpers"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"
)

const DEFAULT_VCHAIN_TIMEOUT = 60 * time.Second

type VChainArgument struct {
	Id       int
	Disabled bool
	BasePort int
}

const basePort = 6000

func (vc VChainArgument) ExternalPort() int {
	port := basePort
	if vc.BasePort != 0 {
		port = vc.BasePort
	}

	return port + vc.Id
}

func managementConfigJson(nodeManagementUrl string, vchainManagementFileUrl string, httpPort int, vChains []VChainArgument) string {
	chains := make([]interface{}, len(vChains))
	model := map[string]interface{}{
		"network": []string{},
		"orchestrator": map[string]interface{}{
			"max-reload-time-delay": "1s",
			"http-port":             httpPort,
			"DynamicManagementConfig": map[string]interface{}{
				"Url":          nodeManagementUrl,
				"ReadInterval": "1m",
				"ResetTimeout": "30m",
			},
		},
		"chains": chains,
		"services": map[string]interface{}{
			"signer": map[string]interface{}{
				"InternalPort": 7777,
				"DockerConfig": map[string]interface{}{
					"Image": "orbsnetwork/signer",
					"Tag":   "experimental",
					"Pull":  false,
				},
			},
		},
	}
	for i, id := range vChains {
		chains[i] = VChainConfig(id, vchainManagementFileUrl)
	}
	jsonStr, _ := json.MarshalIndent(model, "", "    ")
	return string(jsonStr)
}

func VChainConfig(vc VChainArgument, managementFileUrl string) map[string]interface{} {
	return map[string]interface{}{
		"Id":               vc.Id,
		"InternalHttpPort": 8080,
		"InternalPort":     4400,
		"ExternalPort":     vc.ExternalPort(),
		"Disabled":         vc.Disabled,
		"DockerConfig": map[string]interface{}{
			"Image": "orbsnetwork/node",
			"Tag":   "experimental",
			"Pull":  false,
		},
		"Config": map[string]interface{}{
			"active-consensus-algo": 2,
			"management-file-path":  managementFileUrl,
			//"lean-helix-show-debug":                             true,
			//"logger-full-log":                                   true,

			// in case we want to enable benchmark consensus
			//"benchmark-consensus-constant-leader":               genesisValidators[1],
		},
	}
}

func vchainManagementConfig(vcArgument []VChainArgument, topology interface{}, genesisValidator []string) string {
	var committee []map[string]interface{}
	for _, validator := range genesisValidator {
		committee = append(committee, map[string]interface{}{
			"OrbsAddress":  validator,
			"Weight":       1000,
			"IdentityType": 0,
		})
	}

	chains := make(map[string]interface{})
	for _, vc := range vcArgument {
		chains[fmt.Sprintf("%d", vc.Id)] = map[string]interface{}{
			"VirtualChainId":  vc.Id,
			"GenesisRefTime":  0,
			"CurrentTopology": topology,
			"CommitteeEvents": []interface{}{
				map[string]interface{}{
					"RefTime":   0,
					"Committee": committee,
				},
			},
			"SubscriptionEvents": []interface{}{
				map[string]interface{}{
					"RefTime": 0,
					"Data": map[string]interface{}{
						"Status":       "active",
						"Tier":         "B0",
						"RolloutGroup": "main",
						"IdentityType": 0,
						"Params":       make(map[string]interface{}),
					},
				},
			},
			"ProtocolVersionEvents": []interface{}{
				map[string]interface{}{
					"RefTime": 0,
					"Data": map[string]interface{}{
						"RolloutGroup": "main",
						"Version":      1,
					},
				},
			},
		}
	}

	result := map[string]interface{}{
		"CurrentRefTime":   1592834480,
		"PageStartRefTime": 0,
		"PageEndRefTime":   1592834480,
		"VirtualChains":    chains,
	}

	rawJSON, _ := json.MarshalIndent(result, "", "    ")
	return string(rawJSON)
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
