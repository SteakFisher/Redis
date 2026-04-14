package executer

import (
	"fmt"
	"iter"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/SteakFisher/Redis/app/internal/parser"
	"github.com/SteakFisher/Redis/app/internal/store"
)

func Execute(parsed []parser.RESP, conn net.Conn) []byte {
	iterator := slices.Values(parsed)
	next, stop := iter.Pull(iterator)

	Redis := store.Init()

	defer stop()

	if store.SubscribedClients[conn] != nil {
		return subscribedClient(conn, next)
	}

	for {
		parsedValue, valid := next()

		cmd := string(parsedValue.Data)

		switch strings.ToLower(cmd) {
		// Health check cmds
		case "echo":
			parsedValue, valid := next()

			if !valid {
				fmt.Println("No echo message mentioned mentioned")
				return bulk_error()
			}

			return bulk(string(parsedValue.Data))
		case "ping":
			return simple("PONG")

		// General cmds
		case "type":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)

			redisType, err := Redis.Type(key)

			if err != nil {
				fmt.Println(err)

				return simple("none")
			}

			return simple(redisType)

		// Map cmds
		case "set":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No key mentioned in set cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)

			parsedValue, valid := next()

			if !valid {
				fmt.Println("No value mentioned in set cmd")
				return bulk_error()
			}

			value := string(parsedValue.Data)

			parsedValue, valid = next()
			PX := -1

			// Optional parameters, PX EX etc.
			if valid {
				switch strings.ToLower(string(parsedValue.Data)) {
				case "px":
					var err error

					parsedValue, valid := next()

					if !valid {
						fmt.Println("No PX value mentioned in set cmd")
						return bulk_error()
					}

					PX, err = strconv.Atoi(string(parsedValue.Data))

					if err != nil {
						return bulk_error()
					}
				case "ex":
					parsedValue, valid := next()

					if !valid {
						fmt.Println("No EX value mentioned in set cmd")
						return bulk_error()
					}

					EX, err := strconv.Atoi(string(parsedValue.Data))

					if err != nil {
						return bulk_error()
					}

					PX = EX * 1000
				}
			}

			Redis.SetString(key, value, PX)
			return simple("OK")
		case "get":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No key mentioned in get cmd")
				return bulk_error()
			}

			val, err := Redis.Get(string(parsedValue.Data))

			if err != nil {
				fmt.Println(err)
				return bulk_error()
			}

			return bulk(val)

		// List cmds
		case "rpush":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)
			val := make([]string, 0, 4)

			for parsedValue, valid = next(); valid; parsedValue, valid = next() {
				val = append(val, string(parsedValue.Data))
			}

			if len(val) == 0 {
				fmt.Println("No list val provided in rpush cmd")
			}

			arrayLen := Redis.SetArray(key, val, false)

			return integer(arrayLen)
		case "lpush":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)
			val := make([]string, 0, 4)

			for parsedValue, valid = next(); valid; parsedValue, valid = next() {
				val = append([]string{string(parsedValue.Data)}, val...)
			}

			if len(val) == 0 {
				fmt.Println("No list val provided in rpush cmd")
			}

			arrayLen := Redis.SetArray(key, val, true)

			return integer(arrayLen)
		case "lrange":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("Start value not mentioned")
				return bulk_error()
			}

			start, err := strconv.Atoi(string(parsedValue.Data))

			if err != nil {
				fmt.Println("Start range wasn't an int")
			}

			parsedValue, valid = next()

			if !valid {
				fmt.Println("Stop value not mentioned")
				return bulk_error()
			}

			stop, err := strconv.Atoi(string(parsedValue.Data))

			if err != nil {
				fmt.Println("Stop range wasn't an int")
			}

			newArr, err := Redis.Range(key, start, stop)
			fmt.Println(newArr)

			if err != nil {
				fmt.Println(err)
				return bulk_error()
			}

			return array(newArr)
		case "llen":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in llen cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)

			return integer(Redis.Length(key))
		case "lpop":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in lpop cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)
			num := 1

			parsedValue, valid = next()

			if valid {
				var err error
				num, err = strconv.Atoi(string(parsedValue.Data))

				if err != nil {
					fmt.Println("Pop length isn't a number")
					return bulk_error()
				}
			}

			elems, err := Redis.Pop(key, num)

			if err != nil {
				fmt.Println("List key doesn't exist")
				return bulk_error()
			}

			if len(elems.ArrayVal) == 1 {
				return bulk(elems.ArrayVal[0].StringVal)
			} else {
				return array(elems)
			}
		case "blpop":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in blpop cmd")
				return bulk_error()
			}

			key := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in blpop cmd")
				return bulk_error()
			}

			numFloat, err := strconv.ParseFloat(string(parsedValue.Data), 5)
			num := int(numFloat * 1000)

			if err != nil {
				fmt.Printf("Timeout value is not a number.")
				return bulk_error()
			}

			val, err := Redis.BPop(key, num)

			if err != nil {
				fmt.Println(err)
				return null_array()
			}

			return array(val)

		// Stream cmds
		case "xadd":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key mentioned in xadd cmd")
				return bulk_error()
			}

			streamKey := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key value mentioned in xadd cmd")
				return bulk_error()
			}

			entryID := string(parsedValue.Data)

			keyArr := []string{}
			valArr := []string{}

			for parsedValue, valid = next(); valid; parsedValue, valid = next() {
				keyArr = append(keyArr, string(parsedValue.Data))

				parsedValue, valid = next()
				if !valid {
					fmt.Println("No corresponding value for key: ", keyArr[len(keyArr)-1])
					return bulk_error()
				}

				valArr = append(valArr, string(parsedValue.Data))
			}

			id, err := Redis.StreamAdd(streamKey, entryID, keyArr, valArr)

			if err != nil {
				return simple_error(err.Error())
			}

			return bulk(id)
		case "xrange":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key mentioned in xrange cmd")
				return bulk_error()
			}

			streamKey := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key value mentioned in xrange cmd")
				return bulk_error()
			}

			start := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("NO end ID mentioned in xrange cmd")
			}

			end := string(parsedValue.Data)

			return array(Redis.StreamRange(streamKey, start, end))
		case "xread":
			parsedValue, valid = next()

			arg := string(parsedValue.Data)

			switch strings.ToLower(arg) {
			case "streams":
				for _, v := range parsed {
					fmt.Println(string(v.Data))
				}

				var keyArr, idArr []string

				for {
					parsedValue, valid = next()

					if !valid {
						fmt.Println("No stream IDs mentioned in xread streams cmd")
						return bulk_error()
					}

					_, err := strconv.Atoi(strings.Split(string(parsedValue.Data), `-`)[0])

					if err == nil {
						idArr = append(idArr, string(parsedValue.Data))
						break
					} else {
						keyArr = append(keyArr, string(parsedValue.Data))
					}
				}

				for {
					parsedValue, valid = next()

					if !valid {
						break
					}

					idArr = append(idArr, string(parsedValue.Data))
				}

				if len(idArr) != len(keyArr) {
					fmt.Println("Length of both arrays aren't equal")
					return bulk_error()
				}

				return array(Redis.StreamRead(keyArr, idArr))
			case "block":
				parsedValue, valid = next()

				if !valid {
					fmt.Println("Block time not mentioned")
					return bulk_error()
				}

				block, err := strconv.Atoi(string(parsedValue.Data))

				if err != nil {
					fmt.Println("Block time not an int")
					return bulk_error()
				}

				next()

				parsedValue, valid = next()

				if !valid {
					fmt.Println("Key not mentioned in xread cmd")
					return bulk_error()
				}

				key := string(parsedValue.Data)

				parsedValue, valid = next()

				if !valid {
					fmt.Println("Key not mentioned in xread cmd")
					return bulk_error()
				}

				id := string(parsedValue.Data)

				smth, err := Redis.StreamBlockRead(key, id, block)

				if err != nil {
					return null_array()
				}

				return array(smth)

			default:
				fmt.Println("Unknown XREAD Cmd")
				os.Exit(1)
			}

		// Pub Sub cmds
		case "subscribe":
			parsedValue, valid = next()

			if !valid {
				return bulk_error()
			}

			ch := string(parsedValue.Data)

			return array(store.Subscribe(conn, ch))

		default:
			fmt.Println("Unknown Execution Cmd")
			return null_array()
		}
	}
}

func array(arr store.StringArr) []byte {
	if arr.ArrayVal == nil {
		return []byte("*0\r\n")
	}

	final := []byte(fmt.Sprintf("*%d\r\n", len(arr.ArrayVal)))

	for i := 0; i < len(arr.ArrayVal); i++ {
		switch arr.ArrayVal[i].Type {
		case store.String:
			final = append(final, bulk(arr.ArrayVal[i].StringVal)...)
		case store.Integer:
			final = append(final, integer(arr.ArrayVal[i].IntegerVal)...)
		case store.Array:
			final = append(final, array(arr.ArrayVal[i])...)
		}
	}

	return final
}

func bulk(str string) []byte {
	return []byte(fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(str)), str))
}

func bulk_error() []byte {
	return []byte("$-1\r\n")
}

func null_array() []byte {
	return []byte("*-1\r\n")
}

func simple(text string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", text))
}

func integer(text int) []byte {
	val := []byte(fmt.Sprintf(":%d\r\n", text))
	return val
}

func simple_error(text string) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", text))
}
