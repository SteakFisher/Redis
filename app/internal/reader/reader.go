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

		byteString := string(bytes)
		fmt.Println(byteString)

		for i := 0; i < len(byteString); i++ {
			fmt.Printf("%c", byteString[i])
		}

		_, parsedArray := parser.Parse(bytes)

		fmt.Println(parsedArray)

		ret := executer.Execute(parsedArray)

		conn.Write(ret)
	}

}
