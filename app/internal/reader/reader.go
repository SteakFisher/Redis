package reader

import (
	"fmt"
	"net"

	"github.com/SteakFisher/Redis/app/internal/executer"
	"github.com/SteakFisher/Redis/app/internal/parser"
)

func Read(conn net.Conn) {
	bytes := make([]byte, 1000)

	for {
		_, err := conn.Read(bytes)

		if err != nil {
			fmt.Println("Error reading bytes: ", err.Error())
			conn.Close()
			fmt.Println("Connection Closed")
			break
		}

		_, parsedArray := parser.Parse(bytes)

		ret := executer.Execute(parsedArray)

		conn.Write(ret)
	}

}
