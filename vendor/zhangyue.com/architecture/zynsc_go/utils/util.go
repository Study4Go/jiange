// Package utils util
// yanglei
// 2017年05月08日18:20:47
package utils

import (
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

const portEnvName = "ZYAGENT_HTTPPORT"

var (
	localIP string
	port    string
)

// GetIp 获取本机ip地址
func GetIp() (string, error) {
	if localIP != "" {
		return localIP, nil
	}
	conn, err := net.Dial("udp", "10.100.20.32:53")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	return localIP, nil
}

// Random 生成随机数
func Random(num int) int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(num)
}

// 从环境变量获取端口号
func GetPortFromEnv() string {
	if port != "" {
		return port
	}
	return os.Getenv(portEnvName)
}
