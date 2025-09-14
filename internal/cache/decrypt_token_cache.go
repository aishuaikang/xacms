package cache

import "sync"

type DecryptTokenCache interface {
	SetDecryptToken(token string)
	GetDecryptToken() string
}

type decryptTokenCache struct {
	decryptToken      string
	decryptTokenMutex sync.RWMutex
}

func NewDecryptTokenCache() DecryptTokenCache {
	return &decryptTokenCache{
		decryptToken:      "",
		decryptTokenMutex: sync.RWMutex{},
	}
}

// SetDecryptToken 设置解密token
func (c *decryptTokenCache) SetDecryptToken(token string) {
	c.decryptTokenMutex.Lock()
	defer c.decryptTokenMutex.Unlock()
	c.decryptToken = token
}

// GetDecryptToken 获取解密token
func (c *decryptTokenCache) GetDecryptToken() string {
	c.decryptTokenMutex.RLock()
	defer c.decryptTokenMutex.RUnlock()
	return c.decryptToken
}
