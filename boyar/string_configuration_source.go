package boyar

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/strelets"
)

type stringConfigurationSource struct {
	value configValue
}

func NewStringConfigurationSource(input string) (ConfigurationSource, error) {
	return parseStringConfig(input)
}

func (c *stringConfigurationSource) Keys() []byte {
	keys, _ := json.Marshal(c.value.Keys)
	return keys
}

func (c *stringConfigurationSource) Chains() []*strelets.VirtualChain {
	return c.value.Chains
}

func (c *stringConfigurationSource) FederationNodes() []*strelets.FederationNode {
	return c.value.FederationNodes
}
