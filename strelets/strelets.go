package strelets

import (
	"github.com/pkg/errors"
)

type VirtualChainId uint32

type PublicKey string

type Strelets interface {
	GetChain(chain VirtualChainId) (*vchain, error)
	ProvisionVirtualChain(input *ProvisionVirtualChainInput) error
	UpdateFederation(peers map[PublicKey]*Peer)
	RemoveVirtualChain(input *RemoveVirtualChainInput) error
}

type strelets struct {
	vchains map[VirtualChainId]*vchain
	peers   map[PublicKey]*Peer

	root string
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
		return v, errors.Errorf("virtual Chain with id %h not found", chain)
	} else {
		return v, nil
	}
}

func (s *strelets) UpdateFederation(peers map[PublicKey]*Peer) {
	s.peers = peers
}
