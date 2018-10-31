package strelets

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type VirtualChainId uint32

type PublicKey string

type Strelets interface {
	GetChain(chain VirtualChainId) (*vchain, error)
	ProvisionVirtualChain(chain VirtualChainId, configPath string, httpPort int, gossipPort int, dockerConfig *DockerImageConfig) error
	UpdateFederation(peers map[PublicKey]*Peer)
}

type strelets struct {
	vchains map[VirtualChainId]*vchain
	peers   map[PublicKey]*Peer

	root string
}

type vchain struct {
	id           VirtualChainId
	httpPort     int
	gossipPort   int
	dockerConfig *DockerImageConfig
}

type Peer struct {
	IP   string
	Port int
}

func NewStrelets(root string) Strelets {
	return &strelets{
		vchains: make(map[VirtualChainId]*vchain),
		peers:   make(map[PublicKey]*Peer),
		root:    root,
	}
}

func (s *strelets) GetChain(chain VirtualChainId) (*vchain, error) {
	if v, found := s.vchains[chain]; !found {
		return v, errors.Errorf("virtual chain with id %h not found", chain)
	} else {
		return v, nil
	}
}

func (s *strelets) UpdateFederation(peers map[PublicKey]*Peer) {
	s.peers = peers
}

func (s *strelets) ProvisionVirtualChain(chain VirtualChainId, configPath string, httpPort int, gossipPort int, dockerConfig *DockerImageConfig) error {
	v := &vchain{
		id:           chain,
		httpPort:     httpPort,
		gossipPort:   gossipPort,
		dockerConfig: dockerConfig,
	}
	s.vchains[chain] = v

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithVersion("1.38"))
	if err != nil {
		return err
	}

	imageName := v.dockerConfig.FullImageName()
	if v.dockerConfig.Pull {
		pullImage(ctx, cli, imageName)
	}

	containerId, err := s.runContainer(ctx, cli, imageName, v, configPath)
	if err != nil {
		return err
	}

	fmt.Println(containerId)

	return nil
}