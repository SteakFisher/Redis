package executer

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SteakFisher/Redis/app/internal/parser"
)

func Execute(parsed []parser.RESP) []byte {
	i := 0

	for i < len(parsed) {
		fmt.Println(string(parsed[i].Data))
		cmd := string(parsed[i].Data)

		i++
		switch strings.ToLower(cmd) {
		case "echo":
			i++
			return bulk(string(parsed[i-1].Data))
		case "ping":
			return simple_pong()
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
	final := make([]byte, 0)
	final = append(final, '$')
	final = append(final, []byte(strconv.Itoa(len(str)))...)
	final = append(final, '\r')
	final = append(final, '\n')
	final = append(final, []byte(str)...)
	final = append(final, '\r')
	final = append(final, '\n')

	return final
}

func simple_pong() []byte {
	return []byte("+PONG\r\n")
}
