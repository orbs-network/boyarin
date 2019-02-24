package boyar

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/crypto"
	"github.com/orbs-network/boyarin/strelets"
	"github.com/orbs-network/boyarin/strelets/adapter"
)

type NodeConfiguration interface {
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
	OrchestratorOptions() adapter.OrchestratorOptions
	Hash() string

	SetFederationNodes(federationNodes []*strelets.FederationNode) NodeConfiguration
}

type nodeConfiguration struct {
	Chains              []*strelets.VirtualChain    `json:"chains"`
	FederationNodes     []*strelets.FederationNode  `json:"network"`
	OrchestratorOptions adapter.OrchestratorOptions `json:"orchestrator"`
}

type nodeConfigurationContainer struct {
	value nodeConfiguration
}

func (c *nodeConfigurationContainer) Chains() []*strelets.VirtualChain {
	return c.value.Chains
}

func (c *nodeConfigurationContainer) FederationNodes() []*strelets.FederationNode {
	return c.value.FederationNodes
}

func (c *nodeConfigurationContainer) Hash() string {
	data, _ := json.Marshal(c.value)
	return crypto.CalculateHash(data)
}

func (c *nodeConfigurationContainer) OrchestratorOptions() adapter.OrchestratorOptions {
	return c.value.OrchestratorOptions
}

func (c *nodeConfigurationContainer) SetFederationNodes(federationNodes []*strelets.FederationNode) NodeConfiguration {
	c.value.FederationNodes = federationNodes
	return c
}
