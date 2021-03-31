package server

import (
	"encoding/hex"
	"errors"
	"jiange/constant"
	"jiange/redis"

	"github.com/satori/go.uuid"
)

// GenerateApp to initial new app id
func GenerateApp() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	appID := u.String()[0:8]
	if GetSecretKey(appID) != "" {
		return "", errors.New("app id already exists")
	}
	return appID, err
}

// GenerateSecretKey to generate secret key
func GenerateSecretKey() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	}
	buf := make([]byte, 32)
	hex.Encode(buf, u[:])
	return string(buf), err
}

// SaveAppSecret to save appID's secret key
func SaveAppSecret(appID string, secretKey string) (bool, error) {
	return rpool.HSet(constant.AppKeyRedisKey, appID, secretKey)
}
