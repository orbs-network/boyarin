package boyar

import (
	"github.com/orbs-network/boyarin/strelets"
)

type configValue struct {
	Keys map[string]string
	// FIXME: add peers
	Chains []*strelets.VirtualChain `json:"chains"`
}

type ConfigurationSource interface {
	Keys() []byte
	Chains() []*strelets.VirtualChain
}
