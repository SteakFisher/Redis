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

var Pub chan string
var ClientName map[net.Conn]map[string]struct{}
var ChannelClient map[chan string][]net.Conn

var NameChannel map[string]chan string

var redis_store Redis

func Init() Redis {
	if redis_store.m == nil {
		redis_store = Redis{
			m: make(map[string]*RedisValue),
			c: make(map[string]*RedisChan),
		}
	}

	if ClientName == nil {
		ClientName = make(map[net.Conn]map[string]struct{})
	}

	if ChannelClient == nil {
		ChannelClient = make(map[chan string][]net.Conn)
	}

	if NameChannel == nil {
		NameChannel = make(map[string]chan string)
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
