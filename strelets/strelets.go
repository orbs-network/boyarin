package strelets

import "context"

type Strelets interface {
	ProvisionVirtualChain(ctx context.Context, input *ProvisionVirtualChainInput) error
	RemoveVirtualChain(ctx context.Context, input *RemoveVirtualChainInput) error
}

type strelets struct {
	root string
}

func NewStrelets(root string) Strelets {
	return &strelets{
		root: root,
	}
}
