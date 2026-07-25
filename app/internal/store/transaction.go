package store

import (
	"fmt"
	"net"

	"github.com/SteakFisher/Redis/app/internal/parser"
)

func (r Redis) Multi(conn net.Conn) {
	TransactingClients[conn] = [][]parser.RESP{}
}

func (r Redis) QueueTransaction(conn net.Conn, parsed []parser.RESP) {
	TransactingClients[conn] = append(TransactingClients[conn], parsed)
}

func (r Redis) Watch(conn net.Conn, keys []string) {
	for _, v := range keys {
		fmt.Println("key: ", keys)
		WatchedKeys[v] = append(WatchedKeys[v], conn)
		fmt.Println(WatchedKeys)
	}
}
