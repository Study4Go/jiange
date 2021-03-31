package server

import (
	"jiange/constant"
	"jiange/log"
	"jiange/redis"
	"strconv"
	"time"
)

// SaveCount save the count of app accessing the url
func SaveCount(appID, url string) (bool, error) {
	// get access limit
	dayLimit, monthLimit, err := GetAppURLAccessLimit(appID, url)
	if nil != err {
		log.WithFields(log.Fields{
			"app": appID,
			"url": url,
		}).Error("query access limit error")
		return false, err
	}
	_, month, day := time.Now().Date()
	originKey := constant.URLAccessCountKey + url + "_"
	dayKey := originKey + strconv.Itoa(day)
	count, err := rpool.Hincrby(dayKey, appID, 1)
	if err != nil {
		log.WithError(err).Error()
		return false, err
	}
	if count > dayLimit && -1 != dayLimit {
		_, err = rpool.Hincrby(dayKey, appID, -1)
		if nil != err {
			log.WithFields(log.Fields{
				"app": appID,
				"url": url,
			}).Error("decry day access count error")
		}
		return false, nil
	}
	monthKey := originKey + month.String()
	count, err = rpool.Hincrby(monthKey, appID, 1)
	if err != nil {
		log.WithError(err).Error()
		return false, err
	}
	if count > monthLimit && -1 != monthLimit {
		_, err = rpool.Hincrby(monthKey, url, -1)
		if nil != err {
			log.WithFields(log.Fields{
				"app": appID,
				"url": url,
			}).Error("decry month access count error")
		}
		return false, nil
	}
	return true, nil
}
