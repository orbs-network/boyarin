package strelets

import (
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/runconfig"
	"io"
	"os"
	"path/filepath"
)

func pullImage(ctx context.Context, cli *client.Client, imageName string) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
}

func (s *strelets) runContainer(ctx context.Context, cli *client.Client, imageName string, v *vchain, configPath string) (string, error) {
	containerName := getContainerName(v.dockerConfig.Prefix, v.id)

	configDir := createConfigDir(s.root, containerName)
	absolutePathToLogs := createLogsDir(filepath.Join(s.root, "logs"), containerName)
	absoluteNetworkConfigPath := getNetworkConfig(configDir, s.peers)
	absolutePathToConfig, err := copyNodeConfig(configDir, configPath)
	if err != nil {
		return "", err
	}

	jsonConfig, _ := buildJSONConfig(imageName, v.httpPort, v.gossipPort,
		absolutePathToConfig, absolutePathToLogs, absoluteNetworkConfigPath)

	fmt.Println(string(jsonConfig))

	decoder := runconfig.ContainerDecoder{}
	config, hostConfig, networkConfig, err := decoder.DecodeConfig(bytes.NewReader(jsonConfig))
	if err != nil {
		return "", err
	}

	resp, err := cli.ContainerCreate(ctx, config, hostConfig, networkConfig, containerName)
	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}