package server

import (
	"jiange/constant"
	"jiange/log"
	"jiange/redis"
	"strconv"
	"sync"
	"time"
)

const expireDuration = time.Minute // expire time:1 minute

var appSecretConfig configStruct // appID's secret key config:map[appID]secretKey

var namespaceConfig sync.Map // service and namespace config:key is service,value is namespace

var appUrlsConfig sync.Map // appID's support url:key is appID,value is url list

var urlRuleConfig sync.Map // url's rule list:key is url,value is rule list

var urlListConfig configStruct // all url list

var accessLimitConfig sync.Map // access limit count info:key is {appId}_{url},real value is list[day_limit, month_limit]

// config struct
type configStruct struct {
	singleConfig interface{}
	expire       time.Time
	version      string
}

// RouteInfo is route struct containing old and new route info
type RouteInfo struct {
	Old []string
	New []string
}

// GetSecretKey is method to get app's secret key
func GetSecretKey(appID string) string {
	if appID == "" {
		return ""
	}
	// not initial
	if appSecretConfig.singleConfig == nil {
		return refreshAppSecretConfig(appID)
	}
	// expired
	if appSecretConfig.expire.Before(time.Now()) {
		return refreshAppSecretConfig(appID)
	}
	secretKey, ok := appSecretConfig.singleConfig.(map[string]string)[appID]
	// not exist
	if !ok {
		return ""
	}
	return secretKey
}

func refreshAppSecretConfig(appID string) string {
	secretKeyMap, err := rpool.HGetAll(constant.AppKeyRedisKey)
	if err != nil {
		log.WithFields(log.Fields{
			"appID": appID,
		}).Error(err.Error())
		return ""
	}
	expire := time.Now().Add(expireDuration)
	appSecretConfig = configStruct{
		singleConfig: secretKeyMap,
		expire:       expire,
	}
	// todo trace
	secretKey, ok := secretKeyMap[appID]
	if !ok {
		return ""
	}
	return secretKey
}

// GetPlatform by appid
func GetPlatform(appID string) string {
	if appID == "" {
		return ""
	}
	// not initial
	if appSecretConfig.singleConfig == nil {
		return refreshAppPlatformConfig(appID)
	}
	// expired
	if appSecretConfig.expire.Before(time.Now()) {
		return refreshAppPlatformConfig(appID)
	}
	platform, ok := appSecretConfig.singleConfig.(map[string]string)[appID]
	// not exist
	if !ok {
		return ""
	}
	return platform
}

func refreshAppPlatformConfig(appID string) string {
	secretKeyMap, err := rpool.HGetAll(constant.AppKeyRedisKey)
	if err != nil {
		log.WithFields(log.Fields{
			"appID": appID,
		}).Error(err.Error())
		return ""
	}
	expire := time.Now().Add(expireDuration)
	appSecretConfig = configStruct{
		singleConfig: secretKeyMap,
		expire:       expire,
	}
	// todo trace
	platform, ok := secretKeyMap[appID]
	if !ok {
		return ""
	}
	return platform
}

// GetNamespaceConfig is method to get namespace config
func GetNamespaceConfig(url string) (string, error) {
	configInter, ok := namespaceConfig.Load(url)
	// not exist
	if !ok {
		return refreshNamespaceConfig(url)
	}
	config := configInter.(configStruct)
	// expired
	if config.expire.Before(time.Now()) {
		return refreshNamespaceConfig(url)
	}
	return config.singleConfig.(string), nil
}

func refreshNamespaceConfig(url string) (string, error) {
	redisKey := constant.SupportHostRedisKey + url
	namespace, err := rpool.Get(redisKey)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).Error(err.Error())
		return "", err
	}
	expire := time.Now().Add(expireDuration)
	config := configStruct{
		singleConfig: namespace,
		expire:       expire,
	}
	namespaceConfig.Store(url, config)
	return namespace, nil
}

// GetAppUrlsConfig get app and urls config
func GetAppUrlsConfig(appID string) (map[string]string, error) {
	configInter, ok := appUrlsConfig.Load(appID)
	// not exist
	if !ok {
		return refreshAppUrlsConfig(appID)
	}
	config := configInter.(configStruct)
	// expired
	if config.expire.Before(time.Now()) {
		return refreshAppUrlsConfig(appID)
	}
	return config.singleConfig.(map[string]string), nil
}

func refreshAppUrlsConfig(appID string) (map[string]string, error) {
	redisKey := constant.AppUrlsRredisKey + appID
	appUrls, err := rpool.SMembers(redisKey)
	if err != nil || len(appUrls) == 0 {
		log.WithFields(log.Fields{
			"appID": appID,
		}).Error(err.Error())
		return nil, err
	}
	// array to map
	urlMap := make(map[string]string, 0)
	for _, k := range appUrls {
		urlMap[k] = "default"
	}
	expire := time.Now().Add(expireDuration)
	config := configStruct{
		singleConfig: urlMap,
		expire:       expire,
	}
	appUrlsConfig.Store(appID, config)
	return urlMap, nil
}

// GetURLRuleConfig to get url's rule config
func GetURLRuleConfig(url string) (string, error) {
	configInter, ok := urlRuleConfig.Load(url)
	if !ok {
		return refreshURLRuleConfig(url)
	}
	config := configInter.(configStruct)
	if config.expire.Before(time.Now()) {
		return refreshURLRuleConfig(url)
	}
	return config.singleConfig.(string), nil
}

func refreshURLRuleConfig(url string) (string, error) {
	ruleConfig, err := rpool.HGet(constant.URLRuleRedisKey, url)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
		}).Error(err.Error())
		return "", err
	}
	expire := time.Now().Add(expireDuration)
	configStruct := configStruct{
		singleConfig: ruleConfig,
		expire:       expire,
	}
	urlRuleConfig.Store(url, configStruct)
	return ruleConfig, nil
}

// AddURLRuleConfig add URL rule conf
func AddURLRuleConfig(url string, val string) {
	_, err := rpool.HSet(constant.URLRuleRedisKey, url, val)
	if err != nil {
		log.WithFields(log.Fields{
			"url": url,
			"val": val,
		}).Error(err.Error())
	}
}

// GetURLChange to get old and new url list
func GetURLChange() (*RouteInfo, bool) {
	newVersion, err := rpool.Get(constant.URLVersionRedisKey)
	if err != nil {
		log.WithError(err).Error()
		return &RouteInfo{}, false
	}
	// version not changes
	if urlListConfig.version == newVersion {
		return &RouteInfo{}, false
	}
	newList, err := rpool.SMembers(constant.URLListRedisKey)
	if err != nil {
		log.WithFields(log.Fields{}).Error(err.Error())
		return &RouteInfo{}, false
	}
	var oldList []string
	if urlListConfig.singleConfig != nil {
		oldList = urlListConfig.singleConfig.([]string)
	}
	urlListConfig = configStruct{
		singleConfig: newList,
		version:      newVersion,
	}
	return &RouteInfo{
		Old: oldList,
		New: newList,
	}, true
}

// GetAppURLAccessLimit method query the access limit count of app and url
func GetAppURLAccessLimit(appID string, url string) (int, int, error) {
	configInter, ok := accessLimitConfig.Load(appID + "_" + url)
	if !ok {
		return refreshAccessLimit(appID, url)
	}
	config := configInter.(configStruct)
	// expired
	if config.expire.Before(time.Now()) {
		return refreshAccessLimit(appID, url)
	}
	limit := config.singleConfig.([]int)
	return limit[0], limit[1], nil
}

func refreshAccessLimit(appID string, url string) (int, int, error) {
	limit, err := rpool.HGetAll(constant.URLAccessLimitKey + appID + "_" + url)
	if nil != err {
		log.WithError(err).Error()
		return 0, 0, err
	}
	// day limit
	day, err := strconv.Atoi(limit["day"])
	if nil != err {
		log.WithError(err).Error()
		return 0, 0, err
	}
	// month limit
	month, err := strconv.Atoi(limit["month"])
	if nil != err {
		log.WithError(err).Error()
		return 0, 0, err
	}
	configInter := configStruct{
		singleConfig: []int{day, month},
		expire:       time.Now().Add(expireDuration),
	}
	accessLimitConfig.Store(appID+"_"+url, configInter)
	return day, month, nil
}
