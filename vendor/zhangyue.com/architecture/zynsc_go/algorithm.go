// Package zynsc algorithm
// yanglei
// 2017年05月08日18:20:47
package zynsc

import (
	"hash/crc32"

	"zhangyue.com/architecture/zynsc_go/utils"
)

// Algorithm struct
type Algorithm struct {
	Services []*Server
	Path     string
}

// Random 随机
func (a *Algorithm) Random() *Server {
	services := a.filterServiceByURI()
	index := utils.Random(len(services))
	return services[index]
}

// WeightRandom 加权随机
func (a *Algorithm) WeightRandom() *Server {
	services := a.filterServiceByURI()

	var sum int
	for _, service := range services {
		sum += service.Weight
	}
	var service *Server
	if sum > 0 {
		rand := utils.Random(sum)
		for _, v := range services {
			if v.Weight <= 0 {
				continue
			}
			rand -= v.Weight
			if rand < 0 {
				service = v
				break
			}
		}
	}

	return service
}

// SourceHashing 源地址hash
func (a *Algorithm) SourceHashing() *Server {
	services := a.filterServiceByURI()
	source, _ := utils.GetIp()
	sourceInt, _ := utils.Int(crc32.ChecksumIEEE([]byte(source)))
	index := sourceInt % len(services)
	return services[index]
}

// URIService 基于URI过滤服务
func (a *Algorithm) URIService() []*Server {
	return a.filterServiceByURI()
}

// filterServiceByURI 过滤services对象
func (a *Algorithm) filterServiceByURI() []*Server {
	if a.Path != "" {
		var services []*Server
		for _, service := range a.Services {
			// 过滤权重大于0且匹配uri的服务
			if service.URI == a.Path {
				services = append(services, service)
			}
		}

		if len(services) > 0 {
			return services
		}
	}
	return a.Services
}
