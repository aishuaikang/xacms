package cache

import "uav_defender/internal/pkg/config"

type CommonCache interface {
	SetTTL(ttl int64)
	GetTTL() int64
}

type commonCache struct {
	ttl int64
}

func NewCommonCache() CommonCache {
	return &commonCache{
		ttl: config.AppConfig.Configuration.TTL,
	}
}

// SetTTL 设置缓存的TTL
func (c *commonCache) SetTTL(ttl int64) {
	c.ttl = ttl
}

// GetTTL 获取缓存的TTL
func (c *commonCache) GetTTL() int64 {
	return c.ttl
}
