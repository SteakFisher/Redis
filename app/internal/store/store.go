package store

import "fmt"

var redis_store = make(map[string]string)

func Set(key string, val string) {
	redis_store[key] = val

	fmt.Println(redis_store)
}

func Get(key string) (string, error) {
	val := redis_store[key]

	if val == "" {
		return "", fmt.Errorf("Key doesn't exist: %s", key)
	}

	return val, nil
}
