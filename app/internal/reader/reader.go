package reader

import (
	"fmt"
	"net"
	"os"

	"github.com/SteakFisher/Redis/app/internal/config"
	"github.com/SteakFisher/Redis/app/internal/executer"
	"github.com/SteakFisher/Redis/app/internal/parser"
)

func Read(conn net.Conn) {
	Config := config.Default()
	bytesIncoming := make([]byte, 4096)

	for {
		n, err := conn.Read(bytesIncoming)

		if err != nil {

			fmt.Println("Error reading bytes: ", err.Error())
			conn.Close()
			fmt.Println("Connection Closed")
			break
		}

		_, parsedArray := parser.Parse(bytesIncoming[:n])

		ret, isWrite := executer.Execute(parsedArray, conn)

		fmt.Println(isWrite)

		if isWrite {
			baseFileName := Config.Get("dir") + "/" + Config.Get("appenddirname") + "/" + Config.Get("appendfilename")
			appendFileName := baseFileName + ".1.incr.aof"
			appendFile, err := os.OpenFile(appendFileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

			if err != nil {
				fmt.Println("Error opening file", appendFileName)
			}

			defer appendFile.Close()

			// _, err = appendFile.Write(bytes.ReplaceAll(bytesIncoming[:n], []byte("\r\n"), []byte(`\r\n`)))
			_, err = appendFile.Write(bytesIncoming[:n])

			if err != nil {
				fmt.Println(err)
			}

			if Config.Get("appendfsync") == "always" {
				appendFile.Sync()
			}
		}

		conn.Write(ret)
	}

}
