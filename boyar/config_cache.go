package boyar

import "github.com/orbs-network/boyarin/strelets"

type BoyarConfigCache map[string]string

type httpReverseProxyCompositeKey struct {
	Id         strelets.VirtualChainId
	HttpPort   int
	GossipPort int
}

const HTTP_REVERSE_PROXY_HASH = "HTTP_REVERSE_PROXY_HASH"
