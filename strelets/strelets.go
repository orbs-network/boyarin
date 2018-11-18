package strelets

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type Strelets interface {
	ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error
	RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error
}

type strelets struct {
	root         string
	orchestrator adapter.Orchestrator
}

func NewStrelets(root string, docker adapter.Orchestrator) Strelets {
	return &strelets{
		root:         root,
		orchestrator: docker,
	}
}
