package utils

import (
	"encoding/json"
	"github.com/orbs-network/boyarin/crypto"
	"sync"
)

type CacheMap struct {
	mux    sync.Mutex
	values map[string]*CacheFilter
}

func NewCacheMap() *CacheMap {
	return &CacheMap{
		values: make(map[string]*CacheFilter),
	}
}

func (cm *CacheMap) getFilter(key string) *CacheFilter {
	cache, ok := cm.values[key]
	if !ok {
		cm.mux.Lock()
		cache, ok = cm.values[key]
		if !ok {
			cache = NewCacheFilter()
			cm.values[key] = cache
		}
		cm.mux.Unlock()
	}
	return cache
}

func (cm *CacheMap) CheckNewValue(key string, value Hasher) bool {
	return cm.getFilter(key).CheckNewValue(value)
}

func (cm *CacheMap) CheckNewJsonValue(key string, value interface{}) bool {
	return cm.getFilter(key).CheckNewJsonValue(value)
}

func (cm *CacheMap) Clear(key string) {
	cm.mux.Lock()
	delete(cm.values, key)
	cm.mux.Unlock()
}

type CacheFilter struct {
	lastVal string
}

type Hasher interface {
	Hash() string
}

type HashedValue struct {
	Value string
}

func (v *HashedValue) Hash() string {
	return v.Value
}

func NewCacheFilter() *CacheFilter {
	return &CacheFilter{
		lastVal: "init value",
	}
}

func (cf *CacheFilter) checkHash(hash string) bool {
	if hash == cf.lastVal {
		return false
	}
	cf.lastVal = hash
	return true
}

func (cf *CacheFilter) CheckNewValue(value Hasher) bool {
	hash := "hash:" + value.Hash()
	return cf.checkHash(hash)
}

func (cf *CacheFilter) CheckNewJsonValue(value interface{}) bool {
	data, _ := json.Marshal(value)
	hash := "json:" + crypto.CalculateHash(data)
	return cf.checkHash(hash)
}

func (cf *CacheFilter) Clear() {
	cf.lastVal = "cleared"
}
