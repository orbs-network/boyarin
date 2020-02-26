package boyar

import (
	"context"
	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
	"sync"
)

type Cache struct {
	vChains  *utils.CacheMap
	nginx    *utils.CacheFilter
	services *utils.CacheMap
}

func NewCache() *Cache {
	return &Cache{
		vChains:  utils.NewCacheMap(),
		nginx:    utils.NewCacheFilter(),
		services: utils.NewCacheMap(),
	}
}

type Boyar interface {
	ProvisionVirtualChains(ctx context.Context) error
	ProvisionHttpAPIEndpoint(ctx context.Context) error
	ProvisionServices(ctx context.Context) error
}

type boyar struct {
	nginxLock    sync.Mutex
	orchestrator adapter.Orchestrator
	config       config.NodeConfiguration
	cache        *Cache
	logger       log.Logger
}

func NewBoyar(orchestrator adapter.Orchestrator, cfg config.NodeConfiguration, cache *Cache, logger log.Logger) Boyar {
	return &boyar{
		orchestrator: orchestrator,
		config:       cfg,
		cache:        cache,
		logger:       logger,
	}
}
