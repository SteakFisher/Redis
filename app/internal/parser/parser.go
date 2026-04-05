package parser

import (
	"fmt"
	"strconv"
)

const (
	Integer = ':'
	String  = '+'
	Bulk    = '$'
	Array   = '*'
	Error   = '-'
)

type Type = byte

type RESP struct {
	Type  Type
	Raw   []byte
	Data  []byte
	Count int
}

// "*2\r\n$4\r\nECHO\r\n$3\r\nhey\r\n"

func Parse(b []byte) (int, []RESP) {
	var resp RESP

	i := 1
	resp.Type = Type(b[0])
	i += jumpToEOL(b[i:])

	switch resp.Type {
	case Array:
		var err error
		resp.Count, err = strconv.Atoi(string(b[1 : i-2]))

		if err != nil {
			fmt.Println("Error getting resp count")
		}

		respArray := make([]RESP, 0)

		for j := 0; j < resp.Count; j++ {
			n, newresp := Parse(b[i:])
			i += n
			respArray = append(respArray, newresp...)
		}

		return i, respArray
	case Bulk:
		var err error

		resp.Count, err = strconv.Atoi(string(b[1 : i-2]))

		if err != nil {
			fmt.Println("Error getting resp count")
		}

		resp.Raw = b[i : i+resp.Count+2]
		resp.Data = b[i : i+resp.Count]

		respArray := []RESP{resp}
		return i + resp.Count + 2, respArray
	default:
		respArray := []RESP{resp}
		return 8, respArray
	}
}

func jumpToEOL(b []byte) int {
	i := 0
	for {
		if b[i] == '\r' {
			i++
			if b[i] == '\n' {
				i++
				break
			}
		}
		i++
	}
	return i
}
