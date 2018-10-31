package strelets

import (
	"encoding/json"
	"strconv"
)

type DockerImageConfig struct {
	Image  string
	Tag    string
	Pull   bool
	Prefix string
}

func (c *DockerImageConfig) FullImageName() string {
	return c.Image + ":" + c.Tag
}

func getDockerNetworkOptions(httpPort int, gossipPort int) (exposedPorts map[string]interface{}, portBindings map[string][]portBinding) {
	exposedPorts = make(map[string]interface{})
	exposedPorts["8080/tcp"] = struct{}{}
	exposedPorts["4400/tcp"] = struct{}{}

	portBindings = make(map[string][]portBinding)
	portBindings["8080/tcp"] = []portBinding{{"0.0.0.0", strconv.FormatInt(int64(httpPort), 10)}}
	portBindings["4400/tcp"] = []portBinding{{"0.0.0.0", strconv.FormatInt(int64(gossipPort), 10)}}

	return
}

func buildDockerJSONConfig(
	imageName string,
	httpPort int,
	gossipPort int,
	volumes *virtualChainVolumes,
) ([]byte, error) {

	exposedPorts, portBindings := getDockerNetworkOptions(httpPort, gossipPort)

	configMap := make(map[string]interface{})
	configMap["Image"] = imageName
	configMap["ExposedPorts"] = exposedPorts
	configMap["CMD"] = []string{
		"/opt/orbs/orbs-node",
		"--silent",
		"--config", "/opt/orbs/config/node.json",
		"--config", "/opt/orbs/config/network.json",
		"--log", "/opt/orbs/logs/node.log",
	}

	hostConfigMap := make(map[string]interface{})
	hostConfigMap["Binds"] = []string{
		volumes.config + ":/opt/orbs/config/node.json",
		volumes.network + ":/opt/orbs/config/network.json",
		volumes.logs + ":/opt/orbs/logs/",
	}
	hostConfigMap["PortBindings"] = portBindings

	configMap["HostConfig"] = hostConfigMap

	return json.Marshal(configMap)
}

type portBinding struct {
	HostIp   string
	HostPort string
}
