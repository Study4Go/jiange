package rpool

import (
	"jiange/config"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

var pool *redis.Pool
var once sync.Once

func newPool() {
	pool = &redis.Pool{
		MaxIdle:     config.Config.Redis.MaxIdle,
		MaxActive:   config.Config.Redis.MaxActive,
		IdleTimeout: config.Config.Redis.IdleTimeout * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Config.Redis.Address)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
}

func getConnect() redis.Conn {
	once.Do(newPool)
	return pool.Get()
}

// Expire is redis expire method
func Expire(key string, expire int) (int, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("EXPIRE", key, expire)
	if err != nil {
		return 0, err
	}
	return redis.Int(resp, err)
}

// Get method is redis get method
func Get(key string) (string, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("GET", key)
	if err != nil || resp == nil {
		return "", err
	}
	return redis.String(resp, err)
}

// Set method is redis set method
func Set(key string, value string) (bool, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("SET", key, value)
	if err != nil || resp == nil {
		return false, err
	}
	return redis.Bool(resp, err)
}

// HGet method is redis hget method
func HGet(key string, member string) (string, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("HGET", key, member)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", nil
	}
	return redis.String(resp, err)
}

// HSet method is redis hset method
func HSet(key string, member string, val string) (bool, error) {
	conn := getConnect()
	defer conn.Close()
	return redis.Bool(conn.Do("HSET", key, member, val))
}

// HGetAll method is redis hgetall method
func HGetAll(key string) (map[string]string, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("HGETALL", key)
	if err != nil {
		return map[string]string{}, err
	}
	if resp == nil {
		return map[string]string{}, nil
	}
	return redis.StringMap(resp, err)
}

// Hincrby method is redis hincrby method
func Hincrby(key string, member string, count int) (int, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("HINCRBY", key, member, count)
	if err != nil {
		return 0, err
	}
	return redis.Int(resp, err)
}

// SMembers is redis smembers
func SMembers(key string) ([]string, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("SMEMBERS", key)
	if err != nil {
		return []string{}, err
	}
	if resp == nil {
		return []string{}, nil
	}
	return redis.Strings(resp, err)
}

// Lrange is redis lrange method
func Lrange(key string, start int, end int) ([]string, error) {
	conn := getConnect()
	defer conn.Close()
	resp, err := conn.Do("LRANGE", key, start, end)
	if err != nil {
		return []string{}, err
	}
	if resp == nil {
		return []string{}, nil
	}
	return redis.Strings(resp, err)
}
