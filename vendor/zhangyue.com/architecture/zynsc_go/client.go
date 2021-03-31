// Package zynsc client
// yanglei
// 2017年05月08日18:20:47
package zynsc

import (
	"sort"
	"strings"

	"zhangyue.com/architecture/zynsc_go/utils"

	"github.com/golang/groupcache/lru"
)

// A lruCache 对名字空间进行缓存
var lruCache = lru.New(maxNamespaces)

const (
	registerSuffix = ".tcp"
)

// Client struct
type Client struct {
	Port         string
	needRegister bool
}

// NewNSC 创建NSC对象，使用环境变量来获取端口进行注册
// needRegister如果为true则表示必须注册，否则表示无需注册，无需注册的主要是
func NewNSC(needRegister bool) *Client {
	port := ""
	if needRegister {
		port = utils.GetPortFromEnv()
	}
	return &Client{port, needRegister}
}

// MapHostPort2Namespace 读取缓存，将host_port映射成namespace
// param provider string "host_port"
func MapHostPort2Namespace(provider string) (string, bool) {
	return mapHostPort2Namespace(provider)
}

// GetService 获取服务提供者
func (c *Client) GetService(namespace, path, algorithm string) *Server {
	services := c.GetServices(namespace, path)
	algorithmServer := &Algorithm{services, path}

	var server *Server

	switch algorithm {
	case "wr":
		server = algorithmServer.WeightRandom()
	case "sh":
		server = algorithmServer.SourceHashing()
	case "r":
		server = algorithmServer.Random()
	default:
		server = algorithmServer.Random()
	}

	return server
}

// GetMaster 获取主节点服务
func (c *Client) GetMaster(namespace string) *Server {
	return c.GetService(namespace, "/master", "wr")
}

// GetSlave 获取从节点服务列表
func (c *Client) GetSlave(namespace string) []*Server {
	services := c.GetServices(namespace, "/slave")
	algorithmServer := &Algorithm{services, "/slave"}
	return algorithmServer.URIService()
}

// GetServices 批量获取指定服务列表
func (c *Client) GetServices(namespace, path string) []*Server {
	c.checkRegister(namespace)
	nameServer := NewNameServer(namespace)
	services := nameServer.GetServices()
	algorithmServer := &Algorithm{services, path}
	return algorithmServer.URIService()
}

// GetSortedServices 批量获取指定服务列表,服务列表按照有序排列
func (c *Client) GetSortedServices(namespace, path string) []*Server {
	c.checkRegister(namespace)
	nameServer := NewNameServer(namespace)
	services := nameServer.GetServices()
	if len(services) > 1 {
		sort.Sort(ServerSlice(services))
	}
	algorithmServer := &Algorithm{services, path}
	return algorithmServer.URIService()
}

// isRegistered 判断是否注册
func (c *Client) isRegistered(namespace string) bool {
	if _, ok := lruCache.Get(namespace); ok {
		return true
	}
	nameServer := NewNameServer(namespace)
	host, err := utils.GetIp()
	if err != nil {
		return true
	}
	return nameServer.IsConsumerExist(host, c.Port)
}

// register 执行注册消费者
func (c *Client) register(namespace string) bool {
	nameServer := NewNameServer(namespace)
	host, err := utils.GetIp()

	if err != nil {
		return true
	}
	lruCache.Add(namespace, true)
	// 增加校验逻辑，取消对没有端口场景的注册请求
	if c.Port != "" && host != "" {
		return nameServer.AddConsumer(host, c.Port)
	}
	return true
}

// checkRegister 检查是否注册消费者,如果没有注册就进行注册,只注册tcp的名字空间
// tcp的为长连接
func (c *Client) checkRegister(namespace string) {
	if c.needRegister && strings.HasSuffix(namespace, registerSuffix) {
		if !c.isRegistered(namespace) {
			c.register(namespace)
		}
	}
}
