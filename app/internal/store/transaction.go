package store

import (
	"net"

	"github.com/SteakFisher/Redis/app/internal/parser"
)

func (r Redis) Multi(conn net.Conn) {
	TransactingClients[conn] = [][]parser.RESP{}
}

func (r Redis) QueueTransaction(conn net.Conn, parsed []parser.RESP) {
	TransactingClients[conn] = append(TransactingClients[conn], parsed)
}

func (r Redis) Exec(conn net.Conn) ([]byte, error) {
	cmds, _ := TransactingClients[conn]
	defer delete(TransactingClients, conn)

	if len(cmds) == 0 {
		return []byte{}, nil
	}

	return []byte{}, nil
}
