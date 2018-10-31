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
)

func pullImage(ctx context.Context, cli *client.Client, imageName string) {
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
}

func (s *strelets) runContainer(ctx context.Context, cli *client.Client, containerName string, imageName string, v *vchain, vchainVolumes *virtualChainVolumes) (string, error) {
	jsonConfig, _ := buildDockerJSONConfig(imageName, v.httpPort, v.gossipPort, vchainVolumes)

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

func (s *strelets) removeContainer(ctx context.Context, cli *client.Client, containerName string) error {
	return cli.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{
		Force:         true,
		RemoveLinks:   false,
		RemoveVolumes: false,
	})
}
