package boyar

import (
	"github.com/orbs-network/boyarin/strelets"
)

type configValue struct {
	Keys map[string]string
	// FIXME: add peers
	Chains          []*strelets.VirtualChain   `json:"chains"`
	FederationNodes []*strelets.FederationNode `json:"network"`
}

type ConfigurationSource interface {
	Keys() []byte
	FederationNodes() []*strelets.FederationNode
	Chains() []*strelets.VirtualChain
}
