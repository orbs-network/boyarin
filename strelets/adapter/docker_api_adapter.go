package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const DOCKER_API_VERSION = "1.38"

type dockerAPI struct {
	client *client.Client
}

func NewDockerAPI() (DockerAPI, error) {
	client, err := client.NewClientWithOpts(client.WithVersion(DOCKER_API_VERSION))

	if err != nil {
		return nil, err
	}

	return &dockerAPI{client: client}, nil
}

func (d *dockerAPI) PullImage(ctx context.Context, imageName string) error {
	out, err := d.client.ImagePull(ctx, imageName, types.ImagePullOptions{})

	if err != nil {
		return err
	}
	io.Copy(os.Stdout, out)

	return nil
}

func (d *dockerAPI) RunContainer(ctx context.Context, containerName string, dockerConfig interface{}) (string, error) {
	jsonConfig, _ := json.Marshal(dockerConfig)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return "", err
	}

	resp, err := d.client.ContainerCreate(ctx, config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	if err := d.client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (d *dockerAPI) RemoveContainer(ctx context.Context, containerName string) error {
	return d.client.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: false,
	})
}

func (d *dockerAPI) GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int) interface{} {
	exposedPorts, portBindings := buildDockerNetworkOptions(httpPort, gossipPort)
	return buildDockerConfig(imageName, exposedPorts, portBindings, getContainerVolumes(containerName, root))
}

func (d *dockerAPI) StoreConfiguration(ctx context.Context, containerName string, containerRoot string, config *AppConfig) error {
	return storeConfiguration(containerName, containerRoot, config)
}

func storeConfiguration(containerName string, containerRoot string, config *AppConfig) error {
	vchainVolumes := getContainerVolumes(containerName, containerRoot)
	vchainVolumes.createDirs()

	if err := ioutil.WriteFile(vchainVolumes.keyPairConfigFile, config.KeyPair, 0644); err != nil {
		return fmt.Errorf("could not copy key pair config: %s at %s", err, vchainVolumes.keyPairConfigFile)
	}

	if err := ioutil.WriteFile(vchainVolumes.networkConfigFile, config.Network, 0644); err != nil {
		return fmt.Errorf("could not write network config: %s at %s", err, vchainVolumes.networkConfigFile)
	}

	return nil
}

func getContainerVolumes(containerName string, root string) *virtualChainVolumes {
	absolutePathToConfigDir := filepath.Join(root, containerName, "config")
	absolutePathToLogDir, _ := filepath.Abs(filepath.Join(root, containerName, "logs"))

	absolutePathToNetworkConfig, _ := filepath.Abs(filepath.Join(absolutePathToConfigDir, "network.json"))
	absolutePathToKeyPairConfig, _ := filepath.Abs(filepath.Join(absolutePathToConfigDir, "keys.json"))

	return &virtualChainVolumes{
		configRootDir:     absolutePathToConfigDir,
		logsDir:           absolutePathToLogDir,
		keyPairConfigFile: absolutePathToKeyPairConfig,
		networkConfigFile: absolutePathToNetworkConfig,
	}
}

type virtualChainVolumes struct {
	configRootDir string
	logsDir       string

	keyPairConfigFile string
	networkConfigFile string
}

func (v *virtualChainVolumes) createDirs() {
	createDir(v.configRootDir)
	createDir(v.logsDir)
}

func createDir(path string) error {
	return os.MkdirAll(path, 0755)
}

func buildDockerNetworkOptions(httpPort int, gossipPort int) (exposedPorts map[string]interface{}, portBindings map[string][]portBinding) {
	exposedPorts = make(map[string]interface{})
	exposedPorts["8080/tcp"] = struct{}{}
	exposedPorts["4400/tcp"] = struct{}{}

	portBindings = make(map[string][]portBinding)
	portBindings["8080/tcp"] = []portBinding{{"0.0.0.0", strconv.FormatInt(int64(httpPort), 10)}}
	portBindings["4400/tcp"] = []portBinding{{"0.0.0.0", strconv.FormatInt(int64(gossipPort), 10)}}

	return
}

func buildDockerConfig(
	imageName string,
	exposedPorts map[string]interface{},
	portBindings map[string][]portBinding,
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

type portBinding struct {
	HostIp   string
	HostPort string
}
