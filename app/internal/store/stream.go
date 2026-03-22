package store

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (r Redis) StreamAdd(streamKey string, entryID string, keyArr []string, valArr []string) (string, error) {
	val := r.m[streamKey]

	var newStream []map[string]string

	if val == nil {
		val = &RedisValue{
			Type:   Stream,
			Stream: nil,
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

			if val.Stream == nil {
				if milliSecSplit == 0 {
					seqNo = 1
				} else {
					seqNo = 0
				}
			} else {
				lastElem := val.Stream[len(val.Stream)-1]["id"]

				lastIDSplit := strings.Split(lastElem, `-`)

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
		if val.Stream == nil {
			seqNo = 0
		} else {
			lastElem := val.Stream[len(val.Stream)-1]["id"]

			lastIDSplit := strings.Split(lastElem, `-`)

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

	if val.Stream != nil {
		lastElem := val.Stream[len(val.Stream)-1]["id"]

		lastIDSplit := strings.Split(lastElem, `-`)

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

	newEntryID := fmt.Sprintf("%d-%d", milliSecSplit, seqNo)

	entry := map[string]string{
		"id": newEntryID,
	}

	for i, _ := range keyArr {
		entry[keyArr[i]] = valArr[i]
	}

	newStream = append(newStream, entry)

	val.Stream = newStream

	r.m[streamKey] = val

	return newEntryID, nil
}
