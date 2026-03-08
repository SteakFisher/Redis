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

	for i < len(parsed) {
		cmd := string(parsed[i].Data)

		i++
		switch strings.ToLower(cmd) {
		case "echo":
			i++
			return bulk(string(parsed[i-1].Data))
		case "ping":
			return simple("PONG")
		case "set":
			store.Set(string(parsed[i].Data), string(parsed[i+1].Data))
			i += 1
			return simple("OK")
		case "get":
			val, err := store.Get(string(parsed[i].Data))

			if err != nil {
				return bulk("-1")
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

func simple(text string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", text))
}
