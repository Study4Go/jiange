package zynsc

import (
	"fmt"
	"testing"
)

var services = make([]*Server, 0, 3)

func TestWeightRandom(t *testing.T) {
	services = append(services, &Server{Host: "192.168.7.77", Port: "7897", Weight: 0, URI: "/", Disable: 0})
	services = append(services, &Server{"192.168.6.77", "6897", 1, "/", 0})
	services = append(services, &Server{"192.168.5.77", "587", 1, "/", 0})

	algorithm := Algorithm{
		Services: services,
		Path:     "/",
	}

	for i := 0; i < 100; i++ {
		server := algorithm.WeightRandom()
		fmt.Printf("server=%v\n", server)
	}

}
