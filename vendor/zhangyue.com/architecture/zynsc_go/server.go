// Package zynsc server
// yanglei
// 2017年05月08日18:20:47
package zynsc

import (
	"fmt"
	"strings"

	"github.com/bitly/go-simplejson"
)

// Server struct
type Server struct {
	Host    string
	Port    string
	Weight  int
	URI     string
	Disable int
}

// NewServer 构建服务对象
func NewServer(host, port string, jsonConfig string) *Server {
	weight := 1
	uri := "/"
	disable := 0
	ser := &Server{host, port, weight, uri, disable}
	if jsonConfig == "" {
		return ser
	}
	conf, err := simplejson.NewJson([]byte(jsonConfig))
	if err != nil {
		return ser
	}
	weight, err = conf.Get("weight").Int()
	if err == nil {
		ser.Weight = weight
	}
	disable, err = conf.Get("disable").Int()
	if err == nil {
		ser.Disable = disable
	}
	ser.URI = conf.Get("uri").MustString()
	// fmt.Println("URI=", ser.URI, "Weight=", ser.Weight)
	return ser
}

// String 格式化获取http://ip:port
func (s *Server) String() string {
	str := ""
	if s.Host != "" {
		str = fmt.Sprintf("http://%s:%s", s.Host, s.Port)
	}
	return str
}

// ServerSlice 用于支持对Server的slice进行排序操作
type ServerSlice []*Server

// Len 计算长度
func (sl ServerSlice) Len() int {
	return len(sl)
}

// Swap 交换元素
func (sl ServerSlice) Swap(i, j int) {
	sl[i], sl[j] = sl[j], sl[i]
}

// Less 比较大小
func (sl ServerSlice) Less(i, j int) bool {
	return strings.Compare(sl[i].String(), sl[j].String()) < 0
}
