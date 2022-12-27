package boyar

import (
	"context"
	"sync"

	"github.com/orbs-network/boyarin/boyar/config"
	"github.com/orbs-network/boyarin/strelets/adapter"
	"github.com/orbs-network/boyarin/utils"
	"github.com/orbs-network/scribe/log"
)

type Cache struct {
	nginx    *utils.CacheFilter
	services *utils.CacheMap
}

func NewCache() *Cache {
	return &Cache{
		nginx:    utils.NewCacheFilter(),
		services: utils.NewCacheMap(),
	}
}

type Boyar interface {
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
