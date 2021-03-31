package server

import (
	"encoding/json"
	"fmt"
	"jiange/log"
	"jiange/utils"
	"strconv"
	"strings"

	"github.com/caibirdme/yql"
)

// A URLRule url and url confing
type URLRule struct {
	URL    string   `json:"url"`
	Desc   string   `json:"desc"`
	Field  []string `json:"field"`
	RawYQL string   `json:"rawYQL"`
}

// Devingidine api authentication
func Devingidine(url string, appID string) (bool, error) {
	if len(url) == 0 || len(appID) == 0 {
		log.WithFields(log.Fields{
			"url":   url,
			"appID": appID,
		}).Error(fmt.Errorf("Verify params is empty"))
		return false, fmt.Errorf("Verify params is empty")
	}
	return CheckURLByAppID(appID, url)
}

// Verify Parameter calibration
func (g *Gateway) Verify() (bool, error) {
	params := g.RequestParam
	host := g.RequestIP
	url := g.RequestURL
	urlRule := &URLRule{}
	log.WithFields(log.Fields{
		"appid": g.AppID,
	}).Info("Parameters map:", params)
	// get urlrule by url
	urlRule, err := getURLRuleByURL(url)
	if err != nil {
		return false, fmt.Errorf("system error:url rule error")
	}
	if len(urlRule.URL) == 0 {
		log.WithFields(log.Fields{
			"appid": g.AppID,
		}).Info("urlRule result is empty,Verify result is true")
		return true, nil
	}
	// Parameter null check
	err = parameterNilCheck(params, urlRule.Field)
	if err != nil {
		return false, err
	}
	var matchMap map[string]interface{}
	matchMap, err = packagMatchMap(params, urlRule.Field, host)
	if err != nil {
		return false, err
	}
	log.WithFields(log.Fields{
		"appid": g.AppID,
	}).Info("URLRules map:", matchMap)
	result, _ := yql.Match(urlRule.RawYQL, matchMap)
	log.WithFields(log.Fields{
		"appid": g.AppID,
	}).Info("The rules result:", result)
	if !result {
		return false, fmt.Errorf("url rule permission denied")
	}
	return result, nil
}

// Parameter null check
func parameterNilCheck(params map[string][]string, keys []string) error {
	for _, k := range keys {
		splVal := strings.Split(k, "&")[0]
		if len(params[splVal]) == 0 || len(params[splVal][0]) == 0 {
			log.WithFields(log.Fields{
				"parameterNilCheck key": k,
			}).Error("val is empty")
			return fmt.Errorf("key:%s val is empty", k)
		}
	}
	return nil
}

// Parameters are encapsulated
func packagMatchMap(params map[string][]string, keys []string, host string) (map[string]interface{}, error) {
	nk := make(map[string]interface{}, len(keys))
	for _, k := range keys {
		ty := strings.Split(k, "&")[1]
		defaultKey := strings.Split(k, "&")[0]
		intVal, err := strconv.Atoi(ty)
		if err != nil {
			log.WithFields(log.Fields{
				"type":               1,
				"packagMatchMap key": k,
			}).Error(err.Error())
			return nil, err
		}

		if intVal == 1 { // 1:type string
			nk[defaultKey] = string(params[defaultKey][0])
		} else if intVal == 2 { // 2 type int64
			int64Val, err := strconv.ParseInt(params[defaultKey][0], 10, 64)
			if err != nil {
				log.WithFields(log.Fields{
					"type":               2,
					"packagMatchMap key": k,
				}).Error(err.Error())
				return nil, err
			}
			nk[defaultKey] = int64Val
		}
	}
	x := &utils.IntIP{IP: host}
	x.ToIntIP()
	nk["ip"] = x.Intip
	return nk, nil
}

// get URLRules by url
func getURLRuleByURL(url string) (*URLRule, error) {
	val, err := GetURLRuleConfig(url)
	if err != nil {
		log.WithFields(log.Fields{}).Error("GetURLRuleConfig from redis is err:", err.Error())
		return nil, err
	}
	if len(val) == 0 {
		return &URLRule{}, nil
	}
	jsonBlob := []byte(val)
	urlRule := &URLRule{}
	err = json.Unmarshal(jsonBlob, &urlRule)
	if err != nil {
		log.WithFields(log.Fields{}).Error("GetURLRuleConfig err:", err.Error())
		return nil, err
	}
	log.WithFields(log.Fields{}).Info("GetURLRuleConfig result", *urlRule)
	return urlRule, nil
}
