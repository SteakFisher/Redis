package store

import (
	"fmt"
	"slices"
	"time"
)

func (r *Redis) SetArray(key string, val []string, prepend bool) int {
	redisVal := r.m[key]

	var newArr []string

	if redisVal == nil {
		redisVal = &RedisValue{
			Type:  List,
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
		Type:   List,
		Array:  newArr,
		Expiry: time.Time{},
	}

	chanVal := r.c[key]
	fmt.Println(chanVal)

	if chanVal != nil {
		chanVal.mu.Lock()

		ch := chanVal.Array[0]

		if len(chanVal.Array) > 1 {
			chanVal.Array = chanVal.Array[1:]
		} else {
			chanVal.Array = nil
		}

		chanVal.mu.Unlock()

		ch <- 0
	}

	return len(newArr)
}

func (r *Redis) Range(key string, start int, stop int) ([]string, error) {
	val := r.m[key]

	if val == nil {
		val = &RedisValue{
			Type:  List,
			Array: make([]string, 0),
		}
	}

	if val.Array == nil {
		return []string{}, nil
	}

	if val.Type != List {
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
			Type:  List,
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

	if val == nil || len(val.Array) == 0 {
		return []string{}, fmt.Errorf("Key doesn't exist or is empty")
	}

	val.mu.Lock()
	defer val.mu.Unlock()

	newArr := val.Array[num:]
	poppedElems := val.Array[0:num]

	r.m[key] = &RedisValue{
		Type:  List,
		Array: newArr,
	}

	return poppedElems, nil
}

func (r *Redis) BPop(key string, waitTime int) ([]string, error) {
	arr, err := r.Pop(key, 1)

	if err == nil {
		return append([]string{key}, arr...), err
	}

	chanVal := r.c[key]

	if chanVal == nil {
		chanVal = &RedisChan{
			Array: nil,
		}
	}

	chanVal.mu.Lock()

	ch := make(chan int)
	chanVal.Array = append(chanVal.Array, ch)

	r.c[key] = chanVal
	chanVal.mu.Unlock()

	var timerChan <-chan time.Time

	if waitTime > 0 {
		timer := time.NewTimer(time.Millisecond * time.Duration(waitTime))
		timerChan = timer.C
	}

	select {
	case <-ch:
		arr, err = r.Pop(key, 1)

		if err != nil {
			return nil, err
		}

		return append([]string{key}, arr...), err
	case <-timerChan:
		chanVal.mu.Lock()

		chanVal.Array = slices.DeleteFunc(chanVal.Array, func(retChannel chan int) bool {
			if retChannel == ch {
				close(ch)
				return true
			}
			return false
		})

		r.c[key] = chanVal
		chanVal.mu.Unlock()

		return nil, fmt.Errorf("TIMEOUT")
	}
}
