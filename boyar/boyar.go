package boyar

import (
	"github.com/orbs-network/boyarin/strelets"
)

type configValue struct {
	Chains          []*strelets.VirtualChain   `json:"chains"`
	FederationNodes []*strelets.FederationNode `json:"network"`
}

type ConfigurationSource interface {
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
}
