package store

import (
	"fmt"
	"time"
)

func SetArray(key string, val []string, prepend bool) int {
	redisVal := redis_store[key]

	var newArr []string

	if prepend {
		newArr = append(val, redisVal.Array...)
	} else {
		newArr = append(redisVal.Array, val...)
	}

	redis_store[key] = RedisValue{
		Type:   Array,
		Array:  newArr,
		Expiry: time.Time{},
	}

	fmt.Println("Key", key)
	fmt.Print("Len", val)

	return len(newArr)
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

func Length(key string) int {
	val := redis_store[key]

	if val.Array == nil {
		return 0
	}
	return len(val.Array)
}
