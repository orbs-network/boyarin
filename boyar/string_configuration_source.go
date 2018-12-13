package boyar

import (
	"github.com/orbs-network/boyarin/strelets"
)

type stringConfigurationSource struct {
	value nodeConfiguration
	hash  string
}

func NewStringConfigurationSource(input string) (NodeConfiguration, error) {
	return parseStringConfig(input)
}

func (c *stringConfigurationSource) Chains() []*strelets.VirtualChain {
	return c.value.Chains
}

func (c *stringConfigurationSource) FederationNodes() []*strelets.FederationNode {
	return c.value.FederationNodes
}

func (c *stringConfigurationSource) Hash() string {
	return c.hash
}
