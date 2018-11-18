package adapter

import "strconv"

func (d *dockerAPI) GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int, storedConfig interface{}) interface{} {
	exposedPorts, portBindings := buildDockerNetworkOptions(httpPort, gossipPort)
	return buildDockerConfig(imageName, exposedPorts, portBindings, getDockerContainerVolumes(containerName, root))
}

func buildDockerNetworkOptions(httpPort int, gossipPort int) (exposedPorts map[string]interface{}, portBindings map[string][]dockerPortBinding) {
	exposedPorts = make(map[string]interface{})
	exposedPorts["8080/tcp"] = struct{}{}
	exposedPorts["4400/tcp"] = struct{}{}

	portBindings = make(map[string][]dockerPortBinding)
	portBindings["8080/tcp"] = []dockerPortBinding{{"0.0.0.0", strconv.FormatInt(int64(httpPort), 10)}}
	portBindings["4400/tcp"] = []dockerPortBinding{{"0.0.0.0", strconv.FormatInt(int64(gossipPort), 10)}}

	return
}

func buildDockerConfig(
	imageName string,
	exposedPorts map[string]interface{},
	portBindings map[string][]dockerPortBinding,
	volumes *virtualChainVolumes,
) map[string]interface{} {
	configMap := make(map[string]interface{})
	configMap["Image"] = imageName
	configMap["ExposedPorts"] = exposedPorts
	configMap["CMD"] = []string{
		"/opt/orbs/orbs-node",
		"--silent",
		"--config", "/opt/orbs/config/keys.json",
		"--config", "/opt/orbs/config/network.json",
		"--log", "/opt/orbs/logs/node.log",
	}

	hostConfigMap := make(map[string]interface{})
	hostConfigMap["Binds"] = []string{
		volumes.keyPairConfigFile + ":/opt/orbs/config/keys.json",
		volumes.networkConfigFile + ":/opt/orbs/config/network.json",
		volumes.logsDir + ":/opt/orbs/logs/",
	}
	hostConfigMap["PortBindings"] = portBindings

	configMap["HostConfig"] = hostConfigMap

	return configMap
}

type dockerPortBinding struct {
	HostIp   string
	HostPort string
}
