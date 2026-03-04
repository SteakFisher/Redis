package reader

import (
	"fmt"
	"net"
)

func Read(conn net.Conn) {
	bytes := make([]byte, 100)

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

		conn.Write([]byte("+PONG\r\n"))
	}

}
