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
	root   string
	docker adapter.DockerAPI
}

func NewStrelets(root string, docker adapter.DockerAPI) Strelets {
	return &strelets{
		root:   root,
		docker: docker,
	}
}
