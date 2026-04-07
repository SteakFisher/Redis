package store

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

func (r Redis) StreamAdd(streamKey string, entryID string, keyArr []string, valArr []string) (string, error) {
	val := r.m[streamKey]

	if val == nil {
		val = &RedisValue{
			Type:   Stream,
			Stream: StringArr{},
		}
	}

	entrySplit := strings.Split(entryID, `-`)

	if len(entrySplit) == 1 && entrySplit[0] != "*" {
		return "", fmt.Errorf("Malformed entryID")
	}

	var milliSecSplit, seqNo int
	var err1, err2 error

	if len(entrySplit) == 2 {
		if entrySplit[1] == "*" {
			milliSecSplit, err1 = strconv.Atoi(entrySplit[0])

			if err1 != nil {
				return "", fmt.Errorf("Malformed entryID, Not milliseconds")
			}

			if val.Stream.ArrayVal == nil {
				if milliSecSplit == 0 {
					seqNo = 1
				} else {
					seqNo = 0
				}
			} else {
				lastElem := val.Stream.ArrayVal[len(val.Stream.ArrayVal)-2]

				lastIDSplit := strings.Split(lastElem.StringVal, `-`)

				lastIDMilliSecSplit, _ := strconv.Atoi(lastIDSplit[0])
				lastIDSeqNo, _ := strconv.Atoi(lastIDSplit[1])

				if milliSecSplit == lastIDMilliSecSplit {
					seqNo = lastIDSeqNo + 1
				} else {
					seqNo = 0
				}
			}
		} else {
			milliSecSplit, err1 = strconv.Atoi(entrySplit[0])
			seqNo, err2 = strconv.Atoi(entrySplit[1])

			if err1 != nil || err2 != nil {
				return "", fmt.Errorf("Malformed entryID, Not integers")
			}
		}
	} else {
		milliSecSplit = int(time.Now().UnixMilli())
		if val.Stream.ArrayVal == nil {
			seqNo = 0
		} else {
			lastElem := val.Stream.ArrayVal[len(val.Stream.ArrayVal)-2]

			lastIDSplit := strings.Split(lastElem.StringVal, `-`)

			lastIDMilliSecSplit, _ := strconv.Atoi(lastIDSplit[0])
			lastIDSeqNo, _ := strconv.Atoi(lastIDSplit[1])

			if lastIDMilliSecSplit == milliSecSplit {
				seqNo = lastIDSeqNo + 1
			} else {
				seqNo = 0
			}
		}
	}

	if milliSecSplit == 0 && seqNo == 0 {
		return "", fmt.Errorf("ERR The ID specified in XADD must be greater than 0-0")
	}

	if val.Stream.ArrayVal != nil {
		lastElem := val.Stream.ArrayVal[len(val.Stream.ArrayVal)-2]

		lastIDSplit := strings.Split(lastElem.StringVal, `-`)

		lastIDMilliSecSplit, _ := strconv.Atoi(lastIDSplit[0])
		lastIDSeqNo, _ := strconv.Atoi(lastIDSplit[1])

		if milliSecSplit < lastIDMilliSecSplit {
			return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
		} else if milliSecSplit == lastIDMilliSecSplit {
			if seqNo <= lastIDSeqNo {
				return "", fmt.Errorf("ERR The ID specified in XADD is equal or smaller than the target stream top item")
			}
		}
	}

	chanVal := r.c[streamKey]

	if chanVal != nil {
		chanVal.Array[0] <- 0
	}

	val.mu.Lock()
	defer val.mu.Unlock()

	newEntryID := fmt.Sprintf("%d-%d", milliSecSplit, seqNo)

	val.Stream.ArrayVal = append(val.Stream.ArrayVal, StringArr{
		IsString:  true,
		StringVal: newEntryID,
	})

	entryArr := StringArr{
		IsString: false,
		ArrayVal: nil,
	}

	for i := range len(keyArr) {
		entryArr.ArrayVal = append(entryArr.ArrayVal, StringArr{
			IsString:  true,
			StringVal: keyArr[i],
		})

		entryArr.ArrayVal = append(entryArr.ArrayVal, StringArr{
			IsString:  true,
			StringVal: valArr[i],
		})
	}

	val.Stream.ArrayVal = append(val.Stream.ArrayVal, entryArr)

	r.m[streamKey] = val

	return newEntryID, nil
}

func (r Redis) StreamRange(streamKey string, startID string, stopID string) StringArr {
	val := r.m[streamKey]

	var err error

	var startMilli, stopMilli int
	var startSeq, stopSeq int

	if val == nil {
		return StringArr{}
	}

	if startID == "-" {
		startMilli = 0
		startSeq = 0
	} else if stopID == "+" {
		lastElem := strings.Split(val.Stream.ArrayVal[len(val.Stream.ArrayVal)-2].StringVal, `-`)

		stopMilli, _ = strconv.Atoi(lastElem[0])
		stopSeq, _ = strconv.Atoi(lastElem[1])
	}

	if startID != "-" {
		startSplit := strings.Split(startID, `-`)

		startMilli, err = strconv.Atoi(startSplit[0])

		if err != nil {
			fmt.Println("Start ID not integer")
			return StringArr{
				IsString: false,
				ArrayVal: nil,
			}
		}

		if len(startSplit) == 1 {
			startSeq = 0
		} else {
			startSeq, err = strconv.Atoi(startSplit[1])

			if err != nil {
				fmt.Println("Start Seq not integer")
				return StringArr{
					IsString: false,
					ArrayVal: nil,
				}
			}
		}
	}

	if stopID != "+" {
		stopSplit := strings.Split(stopID, `-`)

		stopMilli, err = strconv.Atoi(stopSplit[0])

		if err != nil {
			fmt.Println("Stop ID not integer")
			return StringArr{
				IsString: false,
				ArrayVal: nil,
			}
		}

		if len(stopSplit) == 1 {
			stopSeq = 0

			for i := len(val.Stream.ArrayVal) - 1; i >= 0; i-- {
				if val.Stream.ArrayVal[i].IsString {
					idMilli := strings.Split(val.Stream.ArrayVal[i].StringVal, `-`)

					if idMilli[0] == stopSplit[0] {
						stopSeq, _ = strconv.Atoi(idMilli[1])
						break
					}
				}
			}
		} else {
			stopSeq, err = strconv.Atoi(stopSplit[1])

			if err != nil {
				fmt.Println("Start Seq not integer")
				return StringArr{
					IsString: false,
					ArrayVal: nil,
				}
			}
		}
	}

	if startMilli > stopMilli {
		fmt.Println("Start ID greater than stop")
		return StringArr{}
	}

	finalArr := StringArr{
		IsString: false,
		ArrayVal: nil,
	}

	for i := 0; i < len(val.Stream.ArrayVal); i++ {
		elem := val.Stream.ArrayVal[i]

		if elem.IsString {
			elemSplit := strings.Split(elem.StringVal, `-`)
			elemMilli, _ := strconv.Atoi(elemSplit[0])
			elemSeq, _ := strconv.Atoi(elemSplit[1])

			if elemMilli < startMilli {
				i++
				continue
			} else if elemMilli == startMilli {
				if elemSeq < startSeq {
					i++
					continue
				}
			}

			if elemMilli > stopMilli {
				break
			} else if elemMilli == stopMilli {
				if elemSeq > stopSeq {
					break
				}
			}

			finalArr.ArrayVal = append(finalArr.ArrayVal, StringArr{
				IsString: false,
				ArrayVal: []StringArr{
					elem,
					val.Stream.ArrayVal[i+1],
				},
			})
			i++
		}
	}

	return finalArr
}

func (r Redis) StreamRead(keys []string, startIDs []string) StringArr {
	finalArr := StringArr{
		IsString: false,
		ArrayVal: nil,
	}

	for i, _ := range keys {
		newIDSplit := strings.Split(startIDs[i], `-`)
		num, _ := strconv.Atoi(newIDSplit[1])

		finalArr.ArrayVal = append(finalArr.ArrayVal, StringArr{
			IsString: false,
			ArrayVal: []StringArr{
				{
					IsString:  true,
					StringVal: keys[i],
				},
				r.StreamRange(keys[i], fmt.Sprintf("%s-%d", newIDSplit[0], num+1), "+"),
			},
		})
	}

	return finalArr
}

func (r Redis) StreamBlockRead(streamKey string, id string, blockTime int) (StringArr, error) {
	lastID := r.StreamLast(streamKey)

	if id != "$" {
		stream := r.StreamRead([]string{streamKey}, []string{id})

		if len(stream.ArrayVal[0].ArrayVal) > 1 {
			if stream.ArrayVal[0].ArrayVal[1].ArrayVal != nil {
				return stream, nil
			}
		}
	}

	chanVal := r.c[streamKey]

	if chanVal == nil {
		chanVal = &RedisChan{
			Array: nil,
		}
	}

	chanVal.mu.Lock()

	ch := make(chan int)
	chanVal.Array = append(chanVal.Array, ch)

	r.c[streamKey] = chanVal
	chanVal.mu.Unlock()

	var timerChan <-chan time.Time

	if blockTime > 0 {
		timer := time.NewTimer(time.Millisecond * time.Duration(blockTime))
		timerChan = timer.C
	}

	for {
		select {
		case <-ch:
			if id == "$" {
				return r.StreamRead([]string{streamKey}, []string{lastID}), nil
			}

			stream := r.StreamRead([]string{streamKey}, []string{id})

			if len(stream.ArrayVal[0].ArrayVal) > 1 {
				if stream.ArrayVal[0].ArrayVal[1].ArrayVal != nil {
					return stream, nil
				}
			}

			continue
		case <-timerChan:
			chanVal.mu.Lock()

			chanVal.Array = slices.DeleteFunc(chanVal.Array, func(retChannel chan int) bool {
				if retChannel == ch {
					close(ch)
					return true
				}
				return false
			})

			r.c[streamKey] = chanVal
			chanVal.mu.Unlock()

			return StringArr{
				IsString: false,
				ArrayVal: nil,
			}, fmt.Errorf("TIMEOUT")
		}
	}
}

func (r Redis) StreamLast(key string) string {
	val := r.m[key]

	if val == nil {
		return "0-0"
	}

	return val.Stream.ArrayVal[len(val.Stream.ArrayVal)-2].StringVal
}
