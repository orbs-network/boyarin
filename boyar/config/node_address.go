package config

import "github.com/orbs-network/boyarin/strelets"

func (n *nodeConfigurationContainer) NodeAddress() strelets.NodeAddress {
	cfg, err := n.readKeysConfig()
	if err != nil {
		return "orbs-network"
	}

	return strelets.NodeAddress(cfg.Address())
}
