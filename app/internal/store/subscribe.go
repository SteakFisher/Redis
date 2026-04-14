package store

import "net"

func Subscribe(conn net.Conn, chName string) StringArr {
	channels := SubscribedClients[conn]

	if len(channels) == 0 {
		channels = map[SubscribeChannel]struct{}{}
	}

	channels[SubscribeChannel{
		Name:    chName,
		Channel: make(chan string),
	}] = struct{}{}

	SubscribedClients[conn] = channels

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
				IntegerVal: len(SubscribedClients[conn]),
			},
		},
	}
}
