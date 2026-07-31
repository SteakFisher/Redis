package executer

import (
	"bufio"
	"fmt"
	"iter"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/SteakFisher/Redis/app/internal/config"
	"github.com/SteakFisher/Redis/app/internal/parser"
	"github.com/SteakFisher/Redis/app/internal/store"
)

func Execute(parsed []parser.RESP, conn net.Conn) ([]byte, bool, []string) {
	iterator := slices.Values(parsed)
	next, stop := iter.Pull(iterator)

	Config := config.Default()
	Redis := store.Init()

	defer stop()

	if store.ClientName[conn] != nil {
		return subscribedClient(conn, next), false, nil
	}

	for {
		parsedValue, valid := next()

		cmd := strings.ToLower(string(parsedValue.Data))

		if cmd == "command" {
			return simple(""), true, nil
		}

		_, ok := store.TransactingClients[conn]

		if ok && cmd != "exec" && cmd != "discard" && cmd != "watch" {
			Redis.QueueTransaction(conn, parsed)
			return simple("QUEUED"), false, nil
		} else if ok && cmd == "watch" {
			Discard(conn)
			return simple_error("ERR WATCH inside MULTI is not allowed"), false, nil
		} else if !ok && cmd == "exec" {
			return simple_error("ERR EXEC without MULTI"), false, nil
		} else if !ok && cmd == "discard" {
			return simple_error("ERR DISCARD without MULTI"), false, nil
		}

		switch cmd {
		// Health check cmds
		case "echo":
			parsedValue, valid := next()

			if !valid {
				fmt.Println("No echo message mentioned mentioned")
				return bulk_error(), false, nil
			}

			return bulk(string(parsedValue.Data)), false, nil
		case "ping":
			return simple("PONG"), false, nil

		// General cmds
		case "type":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)

			redisType, err := Redis.Type(key)

			if err != nil {
				fmt.Println(err)

				return simple("none"), false, []string{key}
			}

			return simple(redisType), false, []string{key}

		// Map cmds
		case "set":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No key mentioned in set cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)

			parsedValue, valid := next()

			if !valid {
				fmt.Println("No value mentioned in set cmd")
				return bulk_error(), false, []string{key}
			}

			value := string(parsedValue.Data)

			parsedValue, valid = next()
			PX := -1

			// Optional parameters, PX EX etc.
			if valid {
				switch strings.ToLower(string(parsedValue.Data)) {
				case "px":
					var err error

					parsedValue, valid := next()

					if !valid {
						fmt.Println("No PX value mentioned in set cmd")
						return bulk_error(), false, []string{key}
					}

					PX, err = strconv.Atoi(string(parsedValue.Data))

					if err != nil {
						return bulk_error(), false, []string{key}
					}
				case "ex":
					parsedValue, valid := next()

					if !valid {
						fmt.Println("No EX value mentioned in set cmd")
						return bulk_error(), false, []string{key}
					}

					EX, err := strconv.Atoi(string(parsedValue.Data))

					if err != nil {
						return bulk_error(), false, []string{key}
					}

					PX = EX * 1000
				}
			}

			Redis.SetString(key, value, PX)
			return simple("OK"), true, []string{key}
		case "get":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No key mentioned in get cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)
			val, err := Redis.Get(key)

			if err != nil {
				fmt.Println(err)
				return bulk_error(), false, []string{key}
			}

			return bulk(val), false, []string{key}

		// List cmds
		case "rpush":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)
			val := make([]string, 0, 4)

			for parsedValue, valid = next(); valid; parsedValue, valid = next() {
				val = append(val, string(parsedValue.Data))
			}

			if len(val) == 0 {
				fmt.Println("No list val provided in rpush cmd")
			}

			arrayLen := Redis.SetArray(key, val, false)

			return integer(arrayLen), true, []string{key}
		case "lpush":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)
			val := make([]string, 0, 4)

			for parsedValue, valid = next(); valid; parsedValue, valid = next() {
				val = append([]string{string(parsedValue.Data)}, val...)
			}

			if len(val) == 0 {
				fmt.Println("No list val provided in rpush cmd")
			}

			arrayLen := Redis.SetArray(key, val, true)

			return integer(arrayLen), true, []string{key}
		case "lrange":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in rpush cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("Start value not mentioned")
				return bulk_error(), false, []string{key}
			}

			start, err := strconv.Atoi(string(parsedValue.Data))

			if err != nil {
				fmt.Println("Start range wasn't an int")
			}

			parsedValue, valid = next()

			if !valid {
				fmt.Println("Stop value not mentioned")
				return bulk_error(), false, []string{key}
			}

			stop, err := strconv.Atoi(string(parsedValue.Data))

			if err != nil {
				fmt.Println("Stop range wasn't an int")
			}

			newArr, err := Redis.Range(key, start, stop)
			fmt.Println(newArr)

			if err != nil {
				fmt.Println(err)
				return bulk_error(), false, []string{key}
			}

			return Array(newArr), false, []string{key}
		case "llen":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in llen cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)

			return integer(Redis.Length(key)), false, []string{key}
		case "lpop":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in lpop cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)
			num := 1

			parsedValue, valid = next()

			if valid {
				var err error
				num, err = strconv.Atoi(string(parsedValue.Data))

				if err != nil {
					fmt.Println("Pop length isn't a number")
					return bulk_error(), false, []string{key}
				}
			}

			elems, err := Redis.Pop(key, num)

			if err != nil {
				fmt.Println("List key doesn't exist")
				return bulk_error(), false, []string{key}
			}

			if len(elems.ArrayVal) == 1 {
				return bulk(elems.ArrayVal[0].StringVal), true, []string{key}
			} else {
				return Array(elems), true, []string{key}
			}
		case "blpop":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in blpop cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("No list key mentioned in blpop cmd")
				return bulk_error(), false, []string{key}
			}

			numFloat, err := strconv.ParseFloat(string(parsedValue.Data), 5)
			num := int(numFloat * 1000)

			if err != nil {
				fmt.Printf("Timeout value is not a number.")
				return bulk_error(), false, []string{key}
			}

			val, err := Redis.BPop(key, num)

			if err != nil {
				fmt.Println(err)
				return null_array(), false, []string{key}
			}

			return Array(val), true, []string{key}

		// Stream cmds
		case "xadd":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key mentioned in xadd cmd")
				return bulk_error(), false, nil
			}

			streamKey := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key value mentioned in xadd cmd")
				return bulk_error(), false, []string{streamKey}
			}

			entryID := string(parsedValue.Data)

			keyArr := []string{}
			valArr := []string{}

			for parsedValue, valid = next(); valid; parsedValue, valid = next() {
				keyArr = append(keyArr, string(parsedValue.Data))

				parsedValue, valid = next()
				if !valid {
					fmt.Println("No corresponding value for key: ", keyArr[len(keyArr)-1])
					return bulk_error(), false, []string{streamKey}
				}

				valArr = append(valArr, string(parsedValue.Data))
			}

			id, err := Redis.StreamAdd(streamKey, entryID, keyArr, valArr)

			if err != nil {
				return simple_error(err.Error()), false, []string{streamKey}
			}

			return bulk(id), true, []string{streamKey}
		case "xrange":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key mentioned in xrange cmd")
				return bulk_error(), false, nil
			}

			streamKey := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("No stream key value mentioned in xrange cmd")
				return bulk_error(), false, []string{streamKey}
			}

			start := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("NO end ID mentioned in xrange cmd")
			}

			end := string(parsedValue.Data)

			return Array(Redis.StreamRange(streamKey, start, end)), false, []string{streamKey}
		case "xread":
			parsedValue, valid = next()

			arg := string(parsedValue.Data)

			switch strings.ToLower(arg) {
			case "streams":
				for _, v := range parsed {
					fmt.Println(string(v.Data))
				}

				var keyArr, idArr []string

				for {
					parsedValue, valid = next()

					if !valid {
						fmt.Println("No stream IDs mentioned in xread streams cmd")
						return bulk_error(), false, nil
					}

					_, err := strconv.Atoi(strings.Split(string(parsedValue.Data), `-`)[0])

					if err == nil {
						idArr = append(idArr, string(parsedValue.Data))
						break
					} else {
						keyArr = append(keyArr, string(parsedValue.Data))
					}
				}

				for {
					parsedValue, valid = next()

					if !valid {
						break
					}

					idArr = append(idArr, string(parsedValue.Data))
				}

				if len(idArr) != len(keyArr) {
					fmt.Println("Length of both arrays aren't equal")
					return bulk_error(), false, keyArr
				}

				return Array(Redis.StreamRead(keyArr, idArr)), false, keyArr
			case "block":
				parsedValue, valid = next()

				if !valid {
					fmt.Println("Block time not mentioned")
					return bulk_error(), false, nil
				}

				block, err := strconv.Atoi(string(parsedValue.Data))

				if err != nil {
					fmt.Println("Block time not an int")
					return bulk_error(), false, nil
				}

				next()

				parsedValue, valid = next()

				if !valid {
					fmt.Println("Key not mentioned in xread cmd")
					return bulk_error(), false, nil
				}

				key := string(parsedValue.Data)

				parsedValue, valid = next()

				if !valid {
					fmt.Println("Key not mentioned in xread cmd")
					return bulk_error(), false, []string{key}
				}

				id := string(parsedValue.Data)

				smth, err := Redis.StreamBlockRead(key, id, block)

				if err != nil {
					return null_array(), false, []string{key}
				}

				return Array(smth), false, []string{key}

			default:
				fmt.Println("Unknown XREAD Cmd")
				os.Exit(1)
			}

		// Pub Sub cmds
		case "subscribe":
			parsedValue, valid = next()

			if !valid {
				return bulk_error(), false, nil
			}

			ch := string(parsedValue.Data)

			return Array(store.Subscribe(conn, ch)), false, []string{ch}
		case "publish":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("Channel name not mentioned in publish cmd")
				return bulk_error(), false, nil
			}

			channelName := string(parsedValue.Data)

			parsedValue, valid = next()

			if !valid {
				fmt.Println("Message not mentioned in publish cmd")
				return bulk_error(), false, []string{channelName}
			}

			message := string(parsedValue.Data)

			length, err := store.Publish(channelName, message)

			if err != nil {
				return bulk_error(), false, []string{channelName}
			}

			return integer(length), false, []string{channelName}

		// Config cmds
		case "config":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("Action not mentioned in config cmd")
				return bulk_error(), false, nil
			}

			action := string(parsedValue.Data)

			switch strings.ToLower(action) {
			case "get":
				parsedValue, valid = next()

				if !valid {
					fmt.Println("Option not mentioned in config get cmd")
					return bulk_error(), false, nil
				}

				option := string(parsedValue.Data)
				val := Config.Get(option)

				return Array(store.StringArr{
					Type: store.Array,
					ArrayVal: []store.StringArr{
						{
							Type:      store.String,
							StringVal: option,
						},
						{
							Type:      store.String,
							StringVal: val,
						},
					},
				}), false, nil

			default:
				fmt.Println("Uknown config cmd")
				return null_array(), false, nil
			}

		case "incr":
			parsedValue, valid = next()

			if !valid {
				fmt.Println("Key not mentioned in INCR cmd")
				return bulk_error(), false, nil
			}

			key := string(parsedValue.Data)

			val, err := Redis.Incr(key)

			if err != nil {
				return simple_error(err.Error()), false, []string{key}
			}

			return integer(val), true, []string{key}

		case "multi":
			Redis.Multi(conn)
			return simple("OK"), false, nil
		case "exec":
			_, ok := store.FailTransactionConnections[conn]

			if ok {
				delete(store.TransactingClients, conn)
				delete(store.FailTransactionConnections, conn)
				return null_array(), false, nil
			}

			ret, _ := Exec(conn)

			if len(ret) == 0 {
				return Array(store.StringArr{}), false, nil
			}

			return ret, true, nil
		case "discard":
			Discard(conn)
			return simple("OK"), false, nil

		case "unwatch":
			Redis.Unwatch(conn)
			delete(store.FailTransactionConnections, conn)
			return simple("OK"), false, nil

		case "watch":
			parsedValue, valid := next()

			if !valid {
				fmt.Println("Key not mentioned in WATCH cmd")
				return null_array(), false, nil
			}

			keys := []string{string(parsedValue.Data)}

			for {
				parsedValue, valid = next()

				if !valid {
					break
				}

				keys = append(keys, string(parsedValue.Data))
			}

			Redis.Watch(conn, keys)
			return simple("OK"), false, keys
		default:
			fmt.Println("Unknown Execution Cmd")
			return null_array(), false, nil
		}
	}
}

func BuildAOF() {
	Config := config.Default()

	baseFileName := Config.Get("dir") + "/" + Config.Get("appenddirname") + "/" + Config.Get("appendfilename")

	manifestFileName := baseFileName + ".manifest"
	manifestFile, err := os.Open(manifestFileName)

	if err != nil {
		fmt.Println("Manifest file error", err)
	}
	defer manifestFile.Close()

	scanner := bufio.NewScanner(manifestFile)
	fileName := ""
	for scanner.Scan() {
		words := strings.Fields(scanner.Text())
		flag := false

		for i, word := range words {
			if word == "type" && words[i+1] == "i" {
				flag = true
				break
			}
		}

		if flag == false {
			continue
		}

		for i, word := range words {
			if word == "file" {
				fileName = words[i+1]
				break
			}
		}
	}

	recreationFile, _ := os.Open(Config.Get("dir") + "/" + Config.Get("appenddirname") + "/" + fileName)

	readBytes := make([]byte, 4096)
	n, _ := recreationFile.Read(readBytes)
	count := 0

	var dummyConn net.Conn
	for count < n {
		i, parsedArray := parser.Parse(readBytes[count:n])
		Execute(parsedArray, dummyConn) //nolint:errcheck

		count += i
	}
}

func Array(arr store.StringArr) []byte {
	if arr.ArrayVal == nil {
		return []byte("*0\r\n")
	}

	final := []byte(fmt.Sprintf("*%d\r\n", len(arr.ArrayVal)))

	for i := 0; i < len(arr.ArrayVal); i++ {
		switch arr.ArrayVal[i].Type {
		case store.String:
			final = append(final, bulk(arr.ArrayVal[i].StringVal)...)
		case store.Integer:
			final = append(final, integer(arr.ArrayVal[i].IntegerVal)...)
		case store.Array:
			final = append(final, Array(arr.ArrayVal[i])...)
		}
	}

	return final
}

func bulk(str string) []byte {
	return []byte(fmt.Sprintf("$%s\r\n%s\r\n", strconv.Itoa(len(str)), str))
}

func bulk_error() []byte {
	return []byte("$-1\r\n")
}

func null_array() []byte {
	return []byte("*-1\r\n")
}

func simple(text string) []byte {
	return []byte(fmt.Sprintf("+%s\r\n", text))
}

func integer(text int) []byte {
	val := []byte(fmt.Sprintf(":%d\r\n", text))
	return val
}

func simple_error(text string) []byte {
	return []byte(fmt.Sprintf("-%s\r\n", text))
}
