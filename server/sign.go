package server

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"jiange/constant"
	"jiange/log"

	"sort"
	"strings"
	"time"
)

const (
	limitTime int64 = 300 // duration time limit:300s
)

// SignCheck is method to check the sign from request
func (gateway *Gateway) SignCheck() (bool, error) {
	// get appID's secret key from cache
	secretKey := GetSecretKey(gateway.AppID)
	if secretKey == "" {
		return false, errors.New("null secretKey")
	}
	platform := GetPlatform(constant.Platform + gateway.AppID)
	if platform != "" && platform != "-1" {
		if gateway.Platform == "" {
			return false, errors.New("null platform")
		}
		if gateway.Platform != platform {
			return false, errors.New("error platform")
		}
	}
	// timestamp check
	duration := time.Now().Local().Unix() - gateway.Ts
	if duration < 0 {
		duration = -duration
	}
	if duration > limitTime {
		return false, errors.New("expire request")
	}
	allParams := gateway.FullParams(secretKey)
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
	for i, key := range keys {
		if i != 0 {
			buffer.WriteString("&")
		}
		buffer.WriteString(key)
		buffer.WriteString("=")
		buffer.WriteString(params[key])
	}
	log.WithFields(log.Fields{
		"signBefore": buffer.String(),
	}).Info()
	h := sha256.New()
	h.Write([]byte(buffer.String()))
	return fmt.Sprintf("%x", h.Sum(nil))
}
