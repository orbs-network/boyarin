package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/test/helpers"
	"net"
	"net/http"
	"net/http/httptest"
)

var NETWORK_KEY_CONFIG = []KeyConfig{
	{
		"a328846cd5b4979d68a8c58a9bdfeee657b34de7",
		"901a1a0bfbe217593062a054e561e708707cb814a123474c25fd567a0fe088f8",
	},
	{
		"d27e2e7398e2582f63d0800330010b3e58952ff6",
		"87a210586f57890ae3642c62ceb58f0f0a54e787891054a5a54c80e1da418253",
	},
	{
		"6e2cb55e4cbe97bf5b1e731d51cc2c285d83cbf9",
		"426308c4d11a6348a62b4fdfb30e2cad70ab039174e2e8ea707895e4c644c4ec",
	},
	{
		"c056dfc0d1fbc7479db11e61d1b0b57612bf7f17",
		"1e404ba4e421cedf58dcc3dddcee656569afc7904e209612f7de93e1ad710300",
	},
}

func genesisValidators(keyPairs []KeyConfig) (genesisValidators []string) {
	for _, keyPair := range keyPairs {
		genesisValidators = append(genesisValidators, keyPair.NodeAddress)
	}

	return
}

type KeyConfig struct {
	NodeAddress    string `json:"node-address"`
	NodePrivateKey string `json:"node-private-key,omitempty"` // Very important to omit empty value to produce a valid config
}

type managementServiceConfig struct {
	managementConfig *string
}

func serveConfig(serviceConfig *managementServiceConfig) *httptest.Server {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:0", helpers.LocalIP()))
	if err != nil {
		panic(err)
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/node/management", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, *serviceConfig.managementConfig)
	})

	server := &httptest.Server{
		Listener: l,
		Config: &http.Server{
			Handler: handler,
		},
	}
	server.Start()

	return server
}

// Every vc is on the same shared network -
// a quirk of the shared networks that WILL stop working if shared network name becomes unique
func buildTopology(keyPairs []KeyConfig, vcId int) (topology []interface{}) {
	for _, kp := range keyPairs {
		topology = append(topology, map[string]interface{}{
			"OrbsAddress": kp.NodeAddress,
			"Ip":          fmt.Sprintf("%s-chain-%d", config.NodeAddress(kp.NodeAddress).ShortID(), vcId),
			"Port":        4400,
		})
	}

	return
}

func managementConfigJson(deps boyarDependencies, nodeManagementUrl string) string {

	model := map[string]interface{}{
		"network": []string{},
		"orchestrator": map[string]interface{}{
			"max-reload-time-delay": "1s",
			"http-port":             deps.httpPort,
			"DynamicManagementConfig": map[string]interface{}{
				"Url":          nodeManagementUrl,
				"ReadInterval": "1m",
				"ResetTimeout": "30m",
			},
			"storage-driver":     deps.storageDriver,
			"storage-mount-type": deps.storageMountType,
			"ExecutableImage": map[string]string{
				"Url":    deps.binaryUrl,
				"Sha256": deps.binarySha256,
			},
		},
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

	jsonStr, _ := json.MarshalIndent(model, "", "    ")
	return string(jsonStr)
}
