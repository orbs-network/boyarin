package strelets

import (
	"context"
)

type ProvisionVirtualChainInput struct {
	VirtualChain *VirtualChain
	Peers        *PeersMap
	NodeAddress  NodeAddress

	KeyPairConfig []byte `json:"-"` // Prevents key leak via log
}

type Peer struct {
	IP   string
	Port int
}

type NodeAddress string

type PeersMap map[NodeAddress]*Peer

func (s *strelets) ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error {
	panic("removed")
}
