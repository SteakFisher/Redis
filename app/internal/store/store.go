package store

import (
	"fmt"
	"time"
)

type RedisValue struct {
	Value  string
	Expiry time.Time
}

var redis_store = make(map[string]RedisValue)

func Set(key string, val string, PX int) {
	expiryTime := time.Time{}

	if PX != -1 {
		now := time.Now().UTC()
		expiryTime = now.Add(time.Millisecond * time.Duration(PX))
	}

	redis_store[key] = RedisValue{
		Value:  val,
		Expiry: expiryTime,
	}
}

func Get(key string) (string, error) {
	val := redis_store[key]

	if val.Value == "" {
		return "", fmt.Errorf("Key doesn't exist: %s", key)
	}

	if !val.Expiry.IsZero() {
		compare := time.Now().UTC().Compare(val.Expiry)
		if compare == 1 || compare == 0 {
			delete(redis_store, key)
			return "", fmt.Errorf("Key has expired: %s", key)
		}
	}

	return val.Value, nil
}
