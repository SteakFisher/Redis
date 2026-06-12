package executer

import (
	"fmt"
	"net"

	"github.com/SteakFisher/Redis/app/internal/store"
)

func Exec(conn net.Conn) ([]byte, error) {
	cmds, _ := store.TransactingClients[conn]

	delete(store.TransactingClients, conn)

	if len(cmds) == 0 {
		return []byte{}, nil
	}

	var final [][]byte

	for _, v := range cmds {

		fmt.Println("start")
		for _, s := range v {
			fmt.Println(string(s.Data))
		}
		fmt.Println("end")

		ret, _ := Execute(v, conn)
		final = append(final, ret)
	}

	finalRet := arrayFromBytes(final)

	return finalRet, nil
}

func Discard(conn net.Conn) {
	delete(store.TransactingClients, conn)
}

func arrayFromBytes(arr [][]byte) []byte {
	final := []byte(fmt.Sprintf("*%d\r\n", len(arr)))

	for i := 0; i < len(arr); i++ {
		final = append(final, arr[i]...)
	}

	return final
}
