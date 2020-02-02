package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type Value1 struct {
	Value string
}
type Value2 struct {
	Value string
}

func TestCacheFilterByHash(t *testing.T) {
	var cache = NewCacheFilter()

	assert.True(t, cache.CheckNewValue(&HashedValue{Value: "foo"}))
	assert.False(t, cache.CheckNewValue(&HashedValue{Value: "foo"}))
	assert.True(t, cache.CheckNewValue(&HashedValue{Value: "bar"}))
	assert.False(t, cache.CheckNewValue(&HashedValue{Value: "bar"}))
}

func TestCacheFilterByJson(t *testing.T) {
	var cache = NewCacheFilter()

	assert.True(t, cache.CheckNewJsonValue(&Value1{Value: "foo"}))
	assert.False(t, cache.CheckNewJsonValue(&Value1{Value: "foo"}))
	assert.True(t, cache.CheckNewJsonValue(&Value2{Value: "bar"}))
	assert.False(t, cache.CheckNewJsonValue(&Value1{Value: "bar"}))
}

func TestClearCacheFilter(t *testing.T) {
	var cache = NewCacheFilter()

	assert.True(t, cache.CheckNewValue(&HashedValue{Value: "foo"}))
	cache.Clear()
	assert.True(t, cache.CheckNewValue(&HashedValue{Value: "foo"}))
}

func TestCacheMapByHash(t *testing.T) {
	var cache = NewCacheMap()

	assert.True(t, cache.CheckNewValue("1", &HashedValue{Value: "foo"}))
	assert.False(t, cache.CheckNewValue("1", &HashedValue{Value: "foo"}))
	assert.True(t, cache.CheckNewValue("1", &HashedValue{Value: "bar"}))
	assert.False(t, cache.CheckNewValue("1", &HashedValue{Value: "bar"}))
}

func TestCacheMapByJson(t *testing.T) {
	var cache = NewCacheMap()

	assert.True(t, cache.CheckNewJsonValue("1", &Value1{Value: "foo"}))
	assert.False(t, cache.CheckNewJsonValue("1", &Value1{Value: "foo"}))
	assert.True(t, cache.CheckNewJsonValue("1", &Value2{Value: "bar"}))
	assert.False(t, cache.CheckNewJsonValue("1", &Value1{Value: "bar"}))
}

func TestClearCacheMap(t *testing.T) {
	var cache = NewCacheMap()

	assert.True(t, cache.CheckNewValue("1", &HashedValue{Value: "foo"}))
	cache.Clear("1")
	assert.True(t, cache.CheckNewValue("1", &HashedValue{Value: "foo"}))
}
