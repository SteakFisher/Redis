package store

import (
	"fmt"
	"sync"
	"time"
)

type RedisValueType int

const (
	String RedisValueType = iota
	Array
)

type RedisValue struct {
	mu sync.Mutex

	Type   RedisValueType
	String string
	Array  []string
	Expiry time.Time
}

type Redis struct {
	m map[string]*RedisValue
}

var redis_store Redis

func Init() *Redis {
	if redis_store.m == nil {
		redis_store = Redis{
			m: make(map[string]*RedisValue),
		}
	}

	return &redis_store
}

func (r *Redis) SetString(key string, val string, PX int) {
	expiryTime := time.Time{}

	if PX != -1 {
		now := time.Now().UTC()
		expiryTime = now.Add(time.Millisecond * time.Duration(PX))
	}

	r.m[key] = &RedisValue{
		Type:   String,
		String: val,
		Expiry: expiryTime,
	}
}

func (r *Redis) Get(key string) (string, error) {
	val := r.m[key]

	if val.String == "" {
		return "", fmt.Errorf("Key doesn't exist: %s", key)
	}

	if !val.Expiry.IsZero() {
		compare := time.Now().UTC().Compare(val.Expiry)
		if compare == 1 || compare == 0 {
			delete(r.m, key)
			return "", fmt.Errorf("Key has expired: %s", key)
		}
	}
	return val.String, nil
}
