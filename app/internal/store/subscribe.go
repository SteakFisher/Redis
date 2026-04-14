package store

import (
	"fmt"
	"net"
)

func Subscribe(conn net.Conn, chName string) StringArr {
	channelNames := ClientName[conn]
	channel := NameChannel[chName]

	if channel == nil {
		channel = make(chan string)
	}

	clients := ChannelClient[channel]

	if clients == nil {
		clients = make([]net.Conn, 0)
	}

	clients = append(clients, conn)
	ChannelClient[channel] = clients

	NameChannel[chName] = channel

	if len(channelNames) == 0 {
		channelNames = map[string]struct{}{}
	}

	channelNames[chName] = struct{}{}
	ClientName[conn] = channelNames

	Pub <- chName

	return StringArr{
		Type: Array,
		ArrayVal: []StringArr{
			{
				Type:      String,
				StringVal: "subscribe",
			},
			{
				Type:      String,
				StringVal: chName,
			},
			{
				Type:       Integer,
				IntegerVal: len(ClientName[conn]),
			},
		},
	}
}

func Publish(channelName string, message string) (int, error) {
	channel := NameChannel[channelName]

	if channel == nil {
		return 0, fmt.Errorf("No channel with that name exists")
	}

	channel <- message

	return len(ChannelClient[channel]), nil
}

func Unsubscribe(conn net.Conn, channelName string) StringArr {
	channelNames := ClientName[conn]

	delete(channelNames, channelName)

	channel := NameChannel[channelName]

	clients := ChannelClient[channel]
	var new []net.Conn

	for _, v := range clients {
		if v != conn {
			new = append(new, v)
		}
	}

	ChannelClient[channel] = new

	return StringArr{
		Type: Array,
		ArrayVal: []StringArr{
			{
				Type:      String,
				StringVal: "unsubscribe",
			},
			{
				Type:      String,
				StringVal: channelName,
			},
			{
				Type:       Integer,
				IntegerVal: len(ClientName[conn]),
			},
		},
	}
}
