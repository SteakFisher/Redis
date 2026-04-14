package reader

import (
	"fmt"
	"net"

	"github.com/SteakFisher/Redis/app/internal/executer"
	"github.com/SteakFisher/Redis/app/internal/parser"
)

func Read(conn net.Conn) {
	bytes := make([]byte, 4096)

	for {
		n, err := conn.Read(bytes)

		if err != nil {

			fmt.Println("Error reading bytes: ", err.Error())
			conn.Close()
			fmt.Println("Connection Closed")
			break
		}

		_, parsedArray := parser.Parse(bytes[:n])

		ret := executer.Execute(parsedArray, conn)

		conn.Write(ret)
	}

}
