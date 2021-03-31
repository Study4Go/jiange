// Package zynsc nameserver
// yanglei
// 2017年05月08日18:20:47
package zynsc

import (
	"fmt"
	"strings"
	"sync"
	"zhangyue.com/architecture/zynsc_go/utils"
)

type servicesRecord struct {
	rw             sync.RWMutex
	lastOKservices map[string][]*Server
}

func (s *servicesRecord) Get(namespace string) ([]*Server, bool) {
	s.rw.RLock()
	services, ok := s.lastOKservices[namespace]
	s.rw.RUnlock()
	if ok {
		return services, ok
	}
	return nil, false
}

func (s *servicesRecord) Set(namespace string, services []*Server) {
	s.rw.Lock()
	s.lastOKservices[namespace] = services
	s.rw.Unlock()
}

// mapHostPort2Namespace 读取缓存，将host_port映射成namespace
// param provider string "host_port"
func mapHostPort2Namespace(provider string) (string, bool) {
	return hostPortCache.Get(provider)
}

var (
	servicesCache = servicesRecord{}
	hostPortCache = utils.NewHostPortCache()
)

func init() {
	if servicesCache.lastOKservices == nil {
		servicesCache.lastOKservices = make(map[string][]*Server, 128)
	}
}

// NameServer 名字空间管理
type NameServer struct {
	namespace string
}

// NewNameServer NewNameServer
func NewNameServer(namespace string) *NameServer {
	ser := &NameServer{namespace}
	return ser
}

// GetServices 获取服务列表
func (n *NameServer) GetServices() []*Server {
	providers := getProvidersWithConfig(n.namespace)

	var temp []string
	var services []*Server
	if providers == nil {
		if services, ok := servicesCache.Get(n.namespace); ok {
			return services
		}
		return services
	}
	if len(providers) > 0 {
		for provider, config := range providers {
			temp = strings.Split(provider, "_")
			if len(temp) != 2 {
				continue
			}
			service := NewServer(temp[0], temp[1], config)
			if service.Disable > 0 {
				continue
			}
			if service.Weight > 0 {
				services = append(services, service)
			}
			hostPortCache.Set(n.namespace, provider)
		}
		servicesCache.Set(n.namespace, services)
	} else {
		if services, ok := servicesCache.Get(n.namespace); ok {
			return services
		}
	}
	return services
}

// AddConsumer 注册消费者
func (n *NameServer) AddConsumer(host, port string) bool {
	consumer := host + "_" + port
	return Create(n.namespace, consumer)
}

// IsConsumerExist 判断消费者是否存在
func (n *NameServer) IsConsumerExist(host, port string) bool {
	consumer := host + "_" + port
	consumerPath := getConsumersPath(n.namespace)
	consumers, _ := GetChildren(consumerPath)
	if len(consumers) > 0 {
		for _, v := range consumers {
			if v == consumer {
				return true
			}
		}
	}
	return false
}

// getProvidersWithConfig 批量获取provider以及配置信息
func getProvidersWithConfig(namespace string) map[string]string {
	providersPath := getProvidersPath(namespace)
	//	获取节点
	nodes, err := GetNodes(providersPath)
	if err != nil {
		return nil
	}
	return nodes
}

// getProvidersPath 根据命名空间生成生产者路径
func getProvidersPath(namespace string) string {
	return fmt.Sprintf(providersPrefix, namespace)
}

// getConsumersPath 根据namesapce获取path
func getConsumersPath(namespace string) string {
	return fmt.Sprintf(consumersPrefix, namespace)
}
