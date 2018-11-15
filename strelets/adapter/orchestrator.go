package adapter

import (
	"context"
	"github.com/docker/docker/client"
)

const DOCKER_API_VERSION = "1.38"

type dockerAPI struct {
	client *client.Client
}

type AppConfig struct {
	KeyPair []byte
	Network []byte
}

type Orchestrator interface {
	PullImage(ctx context.Context, imageName string) error
	// TODO replace interface with something else
	GetContainerConfiguration(imageName string, containerName string, root string, httpPort int, gossipPort int) interface{}
	StoreConfiguration(ctx context.Context, containerName string, root string, config *AppConfig) error
	RunContainer(ctx context.Context, containerName string, config interface{}) (string, error)
	RemoveContainer(ctx context.Context, containerName string) error
}
