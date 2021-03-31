// Package zhangyue.com/zynsc_go/utils
// Author fang<fangming@zhangyue.com>
// 2018.3.23
package utils

import "sync"

// hostPortCache 记录host_port到namespace的反查映射
type hostPortCache struct {
	lock  sync.RWMutex
	index map[string]string //index[host_port] = namespace
}

// NewHostPortCache make new instance
func NewHostPortCache() (*hostPortCache) {
	return &hostPortCache{
        index: map[string]string{},
    }
}

// Set 设置映射关系
func (c *hostPortCache) Set(namespace, provider string) {
	c.lock.Lock()
	c.index[provider] = namespace
	c.lock.Unlock()
}

func (c *hostPortCache) Get(provider string) (string, bool) {
	c.lock.RLock()
	services, ok := c.index[provider]
	c.lock.RUnlock()
	if ok {
		return services, ok
	}
	return "", false
}
