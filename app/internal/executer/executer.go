package executer

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SteakFisher/Redis/app/internal/parser"
	"github.com/SteakFisher/Redis/app/internal/store"
)

func Execute(parsed []parser.RESP) []byte {
	i := 0

	fmt.Println(parsed)
	arrayLen := len(parsed)

	for i < arrayLen {
		cmd := string(parsed[i].Data)

		i++
		switch strings.ToLower(cmd) {
		case "echo":
			return bulk(string(parsed[i].Data))
		case "ping":
			return simple("PONG")
		case "set":
			key := string(parsed[i].Data)
			i++
			value := string(parsed[i].Data)
			i++
			PX := -1

			if i < arrayLen {
				switch strings.ToLower(string(parsed[i].Data)) {
				case "px":
					i++
					var err error
					PX, err = strconv.Atoi(string(parsed[i].Data))

					if err != nil {
						return error_bulk()
					}
				case "ex":
					i++
					EX, err := strconv.Atoi(string(parsed[i].Data))

					if err != nil {
						return error_bulk()
					}

					PX = EX * 1000
				}
			}

			store.Set(key, value, PX)
			i += 1
			return simple("OK")
		case "get":
			val, err := store.Get(string(parsed[i].Data))

			if err != nil {
				return error_bulk()
			}

			return bulk(val)
		default:
			fmt.Println("Unknown Execution Cmd")
			os.Exit(1)
		}
	}
	fmt.Println("Unreachable execution code")
	os.Exit(1)
	return []byte{}
}

func bulk(str string) []byte {
	return []byte(fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(str)), str))
}

func error_bulk() []byte {
	return []byte("$-1\r\n")
}

func simple(text string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", text))
}
