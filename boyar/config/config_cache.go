package config

import "github.com/orbs-network/boyarin/strelets"

type BoyarConfigCache map[string]string

type HttpReverseProxyCompositeKey struct {
	Id         strelets.VirtualChainId
	HttpPort   int
	GossipPort int
	Disabled   bool
}

const HTTP_REVERSE_PROXY_HASH = "proxy"
