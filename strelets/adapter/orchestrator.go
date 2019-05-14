package adapter

import (
	"context"
	"io"
	"time"
)

const DOCKER_API_VERSION = "1.40"

const PROXY_CONTAINER_NAME = "http-api-reverse-proxy"

type AppConfig struct {
	KeyPair []byte
	Network []byte
	Config  []byte
}

type ServiceConfig struct {
	Id            uint32
	NodeAddress   string
	ImageName     string
	ContainerName string
	HttpPort      int
	GossipPort    int

	LimitedMemory  int64
	LimitedCPU     float64
	ReservedMemory int64
	ReservedCPU    float64
}

type Runner interface {
	Run(ctx context.Context) error
}

type ContainerStatus struct {
	Name   string
	NodeID string
	State  string
	Error  string

	Logs string

	CreatedAt time.Time
}

type Orchestrator interface {
	PullImage(ctx context.Context, imageName string) error
	Prepare(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) (Runner, error)
	RemoveContainer(ctx context.Context, containerName string) error

	PrepareReverseProxy(ctx context.Context, config *ReverseProxyConfig) (Runner, error)

	GetStatus(ctx context.Context) ([]*ContainerStatus, error)

	io.Closer
}

type OrchestratorOptions struct {
	StorageDriver          string            `json:"storage-driver"`
	StorageOptions         map[string]string `json:"storage-options"`
	MaxReloadTimedDelayStr string            `json:"max-reload-time-delay"`
}

func (s OrchestratorOptions) MaxReloadTimedDelay() time.Duration {
	d, _ := time.ParseDuration(s.MaxReloadTimedDelayStr)
	return d
}

type SSLOptions struct {
	SSLCertificatePath string
	SSLPrivateKeyPath  string
}
