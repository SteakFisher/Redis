package executer

import (
	"fmt"
	"net"
	"strings"

	"github.com/SteakFisher/Redis/app/internal/parser"
	"github.com/SteakFisher/Redis/app/internal/store"
)

func subscribedClient(conn net.Conn, next func() (parser.RESP, bool)) []byte {
	for {
		parsedValue, valid := next()

		if !valid {
			fmt.Println("Subscribe command not found")
			return bulk_error()
		}

		cmd := string(parsedValue.Data)

		switch strings.ToLower(cmd) {
		case "subscribe":
			parsedValue, valid = next()

			if !valid {
				return bulk_error()
			}

			ch := string(parsedValue.Data)

			return Array(store.Subscribe(conn, ch))
		case "unsubscribe":
			parsedValue, valid = next()

			if !valid {
				return bulk_error()
			}

			ch := string(parsedValue.Data)

			return Array(store.Unsubscribe(conn, ch))

		case "psubscribe":
		case "punsubscribe":
		case "ping":
			return Array(store.StringArr{
				Type: store.Array,
				ArrayVal: []store.StringArr{
					{
						Type:      store.String,
						StringVal: "pong",
					},
					{
						Type:      store.String,
						StringVal: "",
					},
				},
			})
		case "quit":
		default:
			return simple_error(
				fmt.Sprintf("ERR Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", strings.ToLower(cmd)),
			)
		}
	}
}
