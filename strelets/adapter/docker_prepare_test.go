package adapter

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getDockerVolumes(t *testing.T) {
	volumes := getVirtualChainDockerContainerVolumes("node1-chain-42", "/tmp")

	require.NotNil(t, volumes)
	require.EqualValues(t, "/tmp/node1-chain-42/config", volumes.configRootDir)
	require.EqualValues(t, "/tmp/node1-chain-42/config/keys.json", volumes.keyPairConfigFile)
	require.EqualValues(t, "/tmp/node1-chain-42/config/network.json", volumes.networkConfigFile)
	require.EqualValues(t, "/tmp/node1-chain-42/config/config.json", volumes.generalConfigFile)
}

const expectedVirtualChainContainerConfig = `{
   "CMD":[
      "/opt/orbs/orbs-node",
      "--silent",
      "--config",
      "/opt/orbs/config/config.json",
      "--config",
      "/opt/orbs/config/keys.json",
      "--config",
      "/opt/orbs/config/network.json",
      "--log",
      "/opt/orbs/logs/node.log"
   ],
   "ExposedPorts":{
      "8080/tcp":{

      }
   },
   "HostConfig":{
      "Binds":[
		 "/tmp/root/v1/config/config.json:/opt/orbs/config/config.json",
         "/tmp/root/v1/config/keys.json:/opt/orbs/config/keys.json",
         "/tmp/root/v1/config/network.json:/opt/orbs/config/network.json"
      ],
      "PortBindings":{
         "8080/tcp":[
            {
               "HostIp":"0.0.0.0",
               "HostPort":"8080"
            }
         ]
      }
   },
   "Image":"orbs:export"
}`

func Test_getVirtualChainContainerConfig(t *testing.T) {
	exposedPorts := make(map[string]interface{})
	exposedPorts["8080/tcp"] = struct{}{}

	portBindings := make(map[string][]dockerPortBinding)
	portBindings["8080/tcp"] = []dockerPortBinding{{"0.0.0.0", "8080"}}

	volumes := &virtualChainVolumes{
		configRootDir:     "/tmp/root/",
		keyPairConfigFile: "/tmp/root/v1/config/keys.json",
		networkConfigFile: "/tmp/root/v1/config/network.json",
		generalConfigFile: "/tmp/root/v1/config/config.json",
	}

	cfg := getVirtualChainContainerConfig("orbs:export", exposedPorts, portBindings, volumes)
	jsonConfig, _ := json.Marshal(cfg)

	require.JSONEq(t, expectedVirtualChainContainerConfig, string(jsonConfig), "expected config does not match generated config")
}
