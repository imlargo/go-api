package cache

import (
	"strconv"

	"github.com/imlargo/go-api/pkg/kv"
)

type CacheKeys struct {
	builder kv.Builder
}

func NewCacheKeys(keyBuilder kv.Builder) *CacheKeys {
	return &CacheKeys{builder: keyBuilder}
}

func (ck *CacheKeys) UserByID(userID uint) string {
	return ck.builder.BuildForEntity("user", strconv.Itoa(int(userID)))
}
