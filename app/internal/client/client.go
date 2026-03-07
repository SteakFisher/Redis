package client

import (
	"fmt"
	"net"
	"os"

	"github.com/SteakFisher/Redis/app/internal/reader"
)

func AcceptClients(l net.Listener, clients []net.Conn) {
	for {
		conn, err := l.Accept()
		fmt.Println("Client Connected")

		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}

		connectionArray := append(clients, conn)

		go reader.Read(conn)

		fmt.Println("Connection Accepted")
		fmt.Println(connectionArray)
	}
}
