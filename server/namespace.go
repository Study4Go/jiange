package server

import (
	"errors"
	"jiange/log"
	"strings"

	zynsc "zhangyue.com/architecture/zynsc_go"
)

var cli *zynsc.Client

// InitialNamespace method create namespace client
func InitialNamespace() {
	cli = zynsc.NewNSC(false)
}

// QueryRealServiceHost to get real service support host
func QueryRealServiceHost(requestURL string) (string, error) {
	// url format check，format：/xxx
	if requestURL == "" || !strings.HasPrefix(requestURL, "/") {
		log.WithFields(log.Fields{
			"requestURL": requestURL,
		}).Error("query real service error")
		return "", errors.New("request url error")
	}
	// get namespace by service name
	namespace, err := GetNamespaceConfig(requestURL)
	if err != nil {
		return "", errors.New("system error:query namespace error")
	}
	return namespaceDeal(namespace)
}

// namespaceDeal get host by namespace
func namespaceDeal(namespace string) (string, error) {
	// not a namespace
	if strings.HasPrefix(namespace, "http") {
		return "", errors.New("namespace format error")
	}
	service := cli.GetService(namespace, "/", "wr")
	if nil == service {
		return "", errors.New("service is nil")
	}
	log.WithFields(log.Fields{
		"namespace": namespace,
		"service":   service.String(),
	}).Info()
	return service.String(), nil

}
