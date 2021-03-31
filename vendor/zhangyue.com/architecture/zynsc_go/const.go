// Package zynsc const
// yanglei
// 2017年05月08日18:20:47
package zynsc

// StochasticAlgorithm 随机算法
var StochasticAlgorithm = map[string]string{
	"random":        "r",
	"weightRandom":  "wr",
	"SourceHashing": "sh",
}

const (
	// zkapiPrefix zookeeper路径
	zkapiPrefix = "/arch_group/zkapi"
	// consumersPrefix 当前服务的消费者服务前缀
	consumersPrefix = zkapiPrefix + "/%s/consumers"
	// providersPrefix 当前服务的生产者服务前缀
	providersPrefix = zkapiPrefix + "/%s/providers"

	// zkapiPath zkapi服务的路径
	zkapiPath = zkapiPrefix + "/arch.zkapi.http/providers"
	// zkapiAddConsumer 添加消费者接口配置
	zkapiAddConsumer = "/v1/zkapi/consumers/%s/%s"
	// maxNamespaces 最大名字空间数
	maxNamespaces = 4096
)
