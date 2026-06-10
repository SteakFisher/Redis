package store

import (
	"fmt"
	"strconv"
)

func (r Redis) Incr(key string) (int, error) {
	val := r.m[key]

	if val == nil || val.String == "" {
		val = &RedisValue{
			Type:   RedisValueType(String),
			String: "0",
		}
	}

	val.mu.Lock()
	defer val.mu.Unlock()

	if val.Type != RedisValueType(String) {
		return 0, fmt.Errorf("ERR value is not an integer or out of range")
	}

	num, err := strconv.Atoi(val.String)

	if err != nil {
		return 0, fmt.Errorf("ERR value is not an integer or out of range")
	}

	val.String = strconv.Itoa(num + 1)

	r.m[key] = val

	return num + 1, nil
}
