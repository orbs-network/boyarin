package config

type NodeAddress string

func (c *nodeConfigurationContainer) NodeAddress() NodeAddress {
	cfg, err := c.readKeysConfig()
	if err != nil {
		return "orbs-network"
	}

	return NodeAddress(cfg.Address())
}

func (n NodeAddress) ShortID() string {
	return string(n)[:6]
}
