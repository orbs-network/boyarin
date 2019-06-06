package strelets

import (
	"context"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type Strelets interface {
	ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error
	RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error
	UpdateReverseProxy(ctx context.Context, input *UpdateReverseProxyInput) error
	UpdateService(ctx context.Context, input *UpdateServiceInput) error
	ProvisionSharedNetwork(ctx context.Context, input *ProvisionSharedNetworkInput) error
}

type strelets struct {
	orchestrator adapter.Orchestrator
}

func NewStrelets(docker adapter.Orchestrator) Strelets {
	return &strelets{
		orchestrator: docker,
	}
}
