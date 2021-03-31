package tmall

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"jiange/log"
	"jiange/server"

	"sort"
	"strings"
)

const (
	limitTime int64 = 300 // duration time limit:300s
)

// SignCheck is method to check the sign from request
func (gateway *TMallGateway) SignCheck() (bool, error) {
	// get appID's secret key from cache
	secretKey := server.GetSecretKey(gateway.AppID)
	if secretKey == "" {
		return false, errors.New("null secretKey")
	}
	allParams := gateway.FullParams()
	signStrAfter := sign(allParams, secretKey)
	log.WithFields(log.Fields{
		"signStrAfter": signStrAfter,
		"signResult":   strings.EqualFold(gateway.Sign, signStrAfter),
	}).Info("sign result(just for jd test)")
	return strings.EqualFold(gateway.Sign, signStrAfter), nil
}

func sign(params map[string]string, secretKey string) string {
	if len(params) == 0 {
		return ""
	}
	var keys []string
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var buffer bytes.Buffer
	buffer.WriteString(secretKey)
	for _, k := range keys {
		buffer.WriteString(k)
		buffer.WriteString(params[k])
	}
	buffer.WriteString(secretKey)
	log.WithFields(log.Fields{
		"MD5SignBefore": buffer.String(),
	}).Info()
	h := md5.New()
	h.Write([]byte(buffer.String()))
	return hex.EncodeToString(h.Sum(nil))
}
