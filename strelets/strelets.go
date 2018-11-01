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

type Peer struct {
	IP   string
	Port int
}

func NewStrelets(root string) Strelets {
	return &strelets{
		root: root,
	}
}
