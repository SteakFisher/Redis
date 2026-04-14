package store

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type RedisValueType int

const (
	StringVal RedisValueType = iota
	List
	Stream
)

type RedisValue struct {
	mu sync.Mutex

	Type RedisValueType

	String string
	Expiry time.Time

	Array []string

	Stream StringArr
}

type RedisChan struct {
	mu sync.Mutex

	Array []chan int
}

type Redis struct {
	m map[string]*RedisValue
	c map[string]*RedisChan
}

type SubscribeChannel struct {
	Name    string
	Channel chan string
}

var SubscribedClients map[net.Conn]map[SubscribeChannel]struct{}

var redis_store Redis

func Init() Redis {
	if redis_store.m == nil {
		redis_store = Redis{
			m: make(map[string]*RedisValue),
			c: make(map[string]*RedisChan),
		}
	}

	if SubscribedClients == nil {
		SubscribedClients = make(map[net.Conn]map[SubscribeChannel]struct{})
	}

	return redis_store
}

func (r *Redis) Type(key string) (string, error) {
	redisVal := r.m[key]

	if redisVal == nil {
		return "", fmt.Errorf("Key doesn't exist")
	}

	switch redisVal.Type {
	case StringVal:
		return "string", nil
	case List:
		return "list", nil
	case Stream:
		return "stream", nil
	default:
		return "", fmt.Errorf("Unknown DataType")
	}
}
