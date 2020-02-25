package config

type NodeAddress string

func (c *nodeConfigurationContainer) NodeAddress() NodeAddress {
	return NodeAddress(c.KeyConfig().Address())
}

func (n NodeAddress) ShortID() string {
	return string(n)[:6]
}
