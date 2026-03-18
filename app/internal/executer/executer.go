package executer

import (
	"fmt"
	"iter"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/SteakFisher/Redis/app/internal/parser"
	"github.com/SteakFisher/Redis/app/internal/store"
)

func Execute(parsed []parser.RESP) []byte {
	iterator := slices.Values(parsed)
	next, stop := iter.Pull(iterator)

	Redis := store.Init()

	defer stop()

	for {
		parsedValue, valid := next()

		// if !valid {
		// 	fmt.Println("No value mentioned")
		// 	return simple("OK")
		// }

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
			val := make([]string, 0)

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
			val := make([]string, 0)

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

			if len(elems) == 1 {
				return bulk(elems[0])
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

			num, err := strconv.Atoi(string(parsedValue.Data))

			if err != nil {
				fmt.Printf("Timeout value is not a number.")
				return bulk_error()
			}

			val, err := Redis.BPop(key, num)

			if err != nil {
				fmt.Println(err)
				return bulk_error()
			}

			return array(val)

		default:
			fmt.Println("Unknown Execution Cmd")
			os.Exit(1)
		}
	}
}

func array(arr []string) []byte {
	if arr == nil {
		return []byte("*0\r\n")
	}

	final := []byte(fmt.Sprintf("*%d\r\n", len(arr)))

	for i := 0; i < len(arr); i++ {
		final = append(final, bulk(arr[i])...)
	}

	return final
}

func bulk(str string) []byte {
	return []byte(fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(str)), str))
}

func bulk_error() []byte {
	return []byte("$-1\r\n")
}

func simple(text string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", text))
}

func integer(text int) []byte {
	val := []byte(fmt.Sprintf(":%d\r\n", text))
	fmt.Println(val, string(val))
	return val
}
