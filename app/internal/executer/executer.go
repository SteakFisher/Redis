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

	defer stop()

	for {
		parsedValue, valid := next()

		if !valid {
			fmt.Println("No value mentioned")
			return bulk_error()
		}

		cmd := string(parsedValue.Data)

		switch strings.ToLower(cmd) {
		case "echo":
			parsedValue, valid := next()

			if !valid {
				fmt.Println("No echo message mentioned mentioned")
				return bulk_error()
			}

			return bulk(string(parsedValue.Data))
		case "ping":
			return simple("PONG")
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

			store.SetString(key, value, PX)

			parsedValue, valid = next()

			if valid {
				fmt.Println("Should've parsed through everything by now")
				return bulk_error()
			} else {
				return simple("OK")
			}
		case "get":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No key mentioned in get cmd")
				return bulk_error()
			}

			val, err := store.Get(string(parsedValue.Data))

			if err != nil {
				return bulk_error()
			}

			return bulk(val)
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

			arrayLen := store.SetArray(key, val)

			return integer(arrayLen)

		default:
			fmt.Println("Unknown Execution Cmd")
			os.Exit(1)
		}
	}
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
