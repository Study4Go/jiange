package server

import (
	"fmt"
	"jiange/log"
)

// A AppIDUrls appid url集合
type AppIDUrls struct {
	AppID string   `json:"appid"`
	Urls  []string `json:"urls"`
}

// CheckURLByAppID 校验appid是否有url权限
func CheckURLByAppID(appID string, path string) (bool, error) {
	appIDUrls, err := GetAppUrlsConfig(appID)
	if err != nil {
		log.WithFields(log.Fields{
			"appID": appID,
		}).Error(err.Error())
		return false, fmt.Errorf("system error:query url")
	}
	val := appIDUrls[path]
	if len(val) != 0 {
		return true, nil
	}
	log.WithFields(log.Fields{
		"appID": appID,
		"path":  path,
	}).Error("not the appid real path from redis")
	return false, fmt.Errorf("permission denied")
}
