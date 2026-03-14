package store

import (
	"fmt"
	"time"
)

type RedisValueType int

const (
	String RedisValueType = iota
	Array
)

type RedisValue struct {
	Type   RedisValueType
	String string
	Array  []string
	Expiry time.Time
}

var redis_store = make(map[string]RedisValue)

func SetString(key string, val string, PX int) {
	expiryTime := time.Time{}

	if PX != -1 {
		now := time.Now().UTC()
		expiryTime = now.Add(time.Millisecond * time.Duration(PX))
	}

	redis_store[key] = RedisValue{
		Type:   String,
		String: val,
		Expiry: expiryTime,
	}
}

func SetArray(key string, val []string) int {
	redisVal := redis_store[key]

	newArr := append(redisVal.Array, val...)

	redis_store[key] = RedisValue{
		Type:   Array,
		Array:  newArr,
		Expiry: time.Time{},
	}

	fmt.Println("Key", key)
	fmt.Print("Len", val)

	return len(newArr)
}

func Get(key string) (string, error) {
	val := redis_store[key]

	if val.String == "" {
		return "", fmt.Errorf("Key doesn't exist: %s", key)
	}

	if !val.Expiry.IsZero() {
		compare := time.Now().UTC().Compare(val.Expiry)
		if compare == 1 || compare == 0 {
			delete(redis_store, key)
			return "", fmt.Errorf("Key has expired: %s", key)
		}
	}

	return val.String, nil
}

func Range(key string, start int, stop int) ([]string, error) {
	val := redis_store[key]

	if val.Array == nil {
		return []string{}, nil
	}

	if val.Type != Array {
		return []string{}, fmt.Errorf("Cannot find range of a non-array")
	}

	if start < 0 {
		start += len(val.Array)
	}
	if stop < 0 {
		stop += len(val.Array)
	}

	if start < 0 {
		start = 0
	}
	if stop < 0 {
		stop = 0
	}

	if start > stop {
		return []string{}, nil
	} else if stop >= len(val.Array) {
		stop = len(val.Array) - 1
	} else if start >= len(val.Array) {
		return []string{}, nil
	}

	stop += 1

	return val.Array[start:stop], nil
}
