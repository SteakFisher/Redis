package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment the code below to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		fmt.Println("Connection Accepted")

		bytes := make([]byte, 100)

		for {
			_, err = conn.Read(bytes)

			if err != nil {
				fmt.Println("Error reading bytes: ", err.Error())
				conn.Close()
				fmt.Println("Connection Closed")
				break
			}

			byteString := string(bytes)
			fmt.Println(byteString)

			// for i := 0; i < len(byteString); i++ {
			// 	fmt.Printf("%c", byteString[i])
			// }

			conn.Write([]byte("+PONG\r\n"))
		}
	}
}
