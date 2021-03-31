// Package zynsc zkserver
// yanglei
// 2017年05月08日18:20:47
package zynsc

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"zhangyue.com/architecture/zynsc_go/qconf"
	"zhangyue.com/architecture/zynsc_go/utils"
)

var idc string

func init() {
	idc = ""
}

// GetChildren 获取子节点
func GetChildren(path string) ([]string, error) {
	keys, err := qconf.GetBatchKeys(path, idc)
	return keys, err
}

// Get 获取节点值
func Get(path string) (string, error) {
	str, err := qconf.GetConf(path, idc)
	return str, err
}

// GetNodes 批量获取节点数据
func GetNodes(path string) (map[string]string, error) {
	nodes, err := qconf.GetBatchConf(path, idc)
	return nodes, err
}

// GetZKAPIConsumerURL 获取zyapi消费url
func GetZKAPIConsumerURL(namespace, consumer string) (string, error) {
	uri := fmt.Sprintf(zkapiAddConsumer, namespace, consumer)

	zkapi, err := getZkAPI()

	if err != nil {
		return "", err
	}

	url := zkapi + uri
	return url, nil
}

// Create 发送请求创建节点
func Create(namespace, consumer string) bool {

	zkAPIURL, err := GetZKAPIConsumerURL(namespace, consumer)
	if err != nil {
		return false
	}

	//	发起http请求
	result, err := httpPost(zkAPIURL)
	if err != nil {
		return false
	}

	if result.StatusCode == 201 {
		return true
	}

	return false
}

// getZkAPI 获取zkapi路径
func getZkAPI() (string, error) {
	zkapis, err := GetChildren(zkapiPath)

	if err != nil {
		return "", err
	}

	index := utils.Random(len(zkapis))
	info := strings.Split(zkapis[index], "_")

	url := "http://" + info[0] + ":" + info[1]

	return url, nil

}

// httpPost http post方法
func httpPost(url string) (*http.Response, error) {
	c := &http.Client{
		Timeout: 3 * time.Second,
	}
	resp, err := c.Post(url, "application/json;charset=utf-8", nil)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
