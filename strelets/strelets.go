package strelets

type VirtualChainId uint32

type PublicKey string

type Strelets interface {
	ProvisionVirtualChain(input *ProvisionVirtualChainInput) error
	RemoveVirtualChain(input *RemoveVirtualChainInput) error
}

type strelets struct {
	root string
}

func NewStrelets(root string) Strelets {
	return &strelets{
		root: root,
	}
}
