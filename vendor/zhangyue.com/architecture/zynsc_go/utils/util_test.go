package utils

import "fmt"
import "testing"

func TestGetIP(t *testing.T) {
	localIP, err := GetIp()
	if err != nil {
		t.Fatal("GetIp test failed")
	} else {
		fmt.Println("localIP =", localIP)
	}
}

// 需要提前设置环境变量export ZYAGENT_HTTPPORT=23456
func TestGetPortFromEnv(t *testing.T) {
	port := GetPortFromEnv()
	if port == "" {
		t.Fatal("GetPortFromEnv failed")
	} else {
		fmt.Println("port =", port)
	}
}
