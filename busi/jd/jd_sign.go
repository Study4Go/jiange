package jd

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"jiange/config"
	"jiange/log"

	"sort"
	"strings"
)

const (
	limitTime int64 = 300 // duration time limit:300s
)

// SignCheck is method to check the sign from request
func (gateway *JDGateway) SignCheck() (bool, error) {
	// get appID's secret key from cache
	secretKey := config.Config.JdSecretKey
	if secretKey == "" {
		return false, errors.New("null secretKey")
	}
	allParams := gateway.FullParams()
	signStrAfter := sign(allParams)
	log.WithFields(log.Fields{
		"signStrAfter": signStrAfter,
		"signResult":   strings.EqualFold(gateway.Sign, signStrAfter),
	}).Info("sign result(just for jd test)")
	return strings.EqualFold(gateway.Sign, signStrAfter), nil
}

func sign(params map[string]string) string {
	if len(params) == 0 {
		return ""
	}
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var buffer bytes.Buffer
	for _, k := range keys {
		if params[k] == "" {
			continue
		}
		buffer.WriteString(k)
		buffer.WriteString(params[k])
	}
	buffer.WriteString(config.Config.JdSecretKey)
	log.WithFields(log.Fields{
		"signBefore": buffer.String(),
	}).Info()
	h := md5.New()
	h.Write([]byte(buffer.String()))
	return hex.EncodeToString(h.Sum(nil))
}
