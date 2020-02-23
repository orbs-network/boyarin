package config

type NodeAddress string

func (n *nodeConfigurationContainer) NodeAddress() NodeAddress {
	cfg, err := n.readKeysConfig()
	if err != nil {
		return "orbs-network"
	}

	return NodeAddress(cfg.Address())
}
