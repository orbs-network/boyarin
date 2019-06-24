package adapter

import (
	"context"
	"io"
	"time"
)

// FIXME revert to 1.40 when Docker 19.03 is released
const DOCKER_API_VERSION = "1.39"

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

	// FIXME rename ports to be less vchain-specific
	HttpPort   int
	GossipPort int

	LimitedMemory  int64
	LimitedCPU     float64
	ReservedMemory int64
	ReservedCPU    float64

	BlocksVolumeSize int
	LogsVolumeSize   int
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

	PrepareService(ctx context.Context, serviceConfig *ServiceConfig, appConfig *AppConfig) (Runner, error)
	PrepareReverseProxy(ctx context.Context, config *ReverseProxyConfig) (Runner, error)

	GetOverlayNetwork(ctx context.Context, name string) (string, error)

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
