package config

import (
	"github.com/orbs-network/boyarin/strelets"
	"sync"
)

type HttpReverseProxyCompositeKey struct {
	Id         strelets.VirtualChainId
	HttpPort   int
	GossipPort int
	Disabled   bool
}

const HTTP_REVERSE_PROXY_HASH = "proxy"

type Cache interface {
	Get(key string) string
	Put(key string, value string)
	Remove(key string)
}

func NewCache() Cache {
	return &cache{}
}

type cache struct {
	sync.Map
}

func (c *cache) Get(key string) string {
	if value, ok := c.Load(key); ok {
		return value.(string)
	}

	return ""
}

func (c *cache) Put(key string, value string) {
	c.Store(key, value)
}

func (c *cache) Remove(key string) {
	c.Delete(key)
}
