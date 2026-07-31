package store

import (
	"net"
	"slices"

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
		WatchedKeys[v] = append(WatchedKeys[v], conn)
		WatchingConnections[conn] = append(WatchingConnections[conn], v)
	}
}

func (r Redis) Unwatch(conn net.Conn) {
	for _, v := range WatchingConnections[conn] {
		WatchedKeys[v] = slices.DeleteFunc(WatchedKeys[v], func(e net.Conn) bool {
			return e == conn
		})
	}

	WatchingConnections[conn] = []string{}
}
