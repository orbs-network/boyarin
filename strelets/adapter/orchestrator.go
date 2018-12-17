package adapter

import (
	"context"
	"github.com/docker/docker/client"
)

const DOCKER_API_VERSION = "1.38"

const PROXY_CONTAINER_NAME = "http-api-reverse-proxy"

type dockerAPI struct {
	client *client.Client
	root   string
}

type AppConfig struct {
	KeyPair []byte
	Network []byte
}

type Runner interface {
	Run(ctx context.Context) error
}

type Orchestrator interface {
	PullImage(ctx context.Context, imageName string) error
	Prepare(ctx context.Context, imageName string, containerName string, httpPort int, gossipPort int, config *AppConfig) (Runner, error)
	RemoveContainer(ctx context.Context, containerName string) error

	PrepareReverseProxy(ctx context.Context, config string) (Runner, error)

	Close() error
}
