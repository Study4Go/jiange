// 测试client相关接口
package zynsc

import (
	"testing"
)

var cli = NewNSC(true)

func TestGetService(t *testing.T) {
	service := cli.GetService("arch.archapi.http", "/", "wr")
	if service == nil {
		t.Fatal(`"GetService("arch.arcapi.http") failed. Got nil, expected 192.168.6.92_8300"`)
	}
}

func TestGetServices(t *testing.T) {
	services := cli.GetServices("arch.archapi.http", "/")
	if len(services) == 0 {
		t.Fatal(`"GetServices("arch.arcapi.http") failed. Got [], expected ["192.168.6.92_8300"]"`)
	}
}

func TestGetSortedServices(t *testing.T) {
	services1 := cli.GetSortedServices("arch.test.http", "/")
	services2 := cli.GetSortedServices("arch.test.http", "/")
	if len(services1) == 0 {
		t.Fatal(`"GetSortedServices("arch.test.http") failed. Got []"`)
	} else {
		for k, v := range services1 {
			if services2[k].String() != v.String() {
				t.Fatal(`"GetSortedServices("arch.test.http") failed. Expected sorted"`)
			}
		}
	}
}

func TestGetMaster(t *testing.T) {
	masterService := cli.GetMaster("arch.archapi_mysql.tcp")
	if masterService == nil {
		t.Fatal(`"GetMaster("arch.arcapi.http") failed. Got nil"`)
	}
}

func TestGetSlave(t *testing.T) {
	slaveServices := cli.GetSlave("arch.archapi_wlc.http")
	if len(slaveServices) == 0 {
		t.Fatal(`"GetSlave("arch.arcapi.http") failed. Got []"`)
	}
}

func BenchmarkGetService(b *testing.B) {
	for n := 0; n < b.N; n++ {
		cli.GetService("arch.test.http", "/", "wr")
	}
}
