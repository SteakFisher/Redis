package store

import (
	"fmt"
	"time"
)

func (r *Redis) SetArray(key string, val []string, prepend bool) int {
	redisVal := r.m[key]

	var newArr []string

	if redisVal == nil {
		redisVal = &RedisValue{
			Type:  Array,
			Array: make([]string, 0),
		}
	}

	redisVal.mu.Lock()
	defer redisVal.mu.Unlock()

	if prepend {
		newArr = append(val, redisVal.Array...)
	} else {
		newArr = append(redisVal.Array, val...)
	}

	r.m[key] = &RedisValue{
		Type:   Array,
		Array:  newArr,
		Expiry: time.Time{},
	}

	return len(newArr)
}

func (r *Redis) Range(key string, start int, stop int) ([]string, error) {
	val := r.m[key]

	if val == nil {
		val = &RedisValue{
			Type:  Array,
			Array: make([]string, 0),
		}
	}

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

func (r Redis) Length(key string) int {
	val := r.m[key]

	if val == nil {
		val = &RedisValue{
			Type:  Array,
			Array: make([]string, 0),
		}
	}

	if val.Array == nil {
		return 0
	}
	return len(val.Array)
}

func (r *Redis) Pop(key string, num int) ([]string, error) {
	val := r.m[key]

	if val == nil {
		val = &RedisValue{
			Type:  Array,
			Array: make([]string, 0),
		}
	}

	val.mu.Lock()
	defer val.mu.Unlock()

	if val.Array == nil {
		return []string{}, fmt.Errorf("Key doesn't exist")
	}

	newArr := val.Array[num:]
	poppedElems := val.Array[0:num]

	r.m[key] = &RedisValue{
		Type:  Array,
		Array: newArr,
	}

	return poppedElems, nil
}
