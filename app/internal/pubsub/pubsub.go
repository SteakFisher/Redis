package pubsub

import (
	"github.com/SteakFisher/Redis/app/internal/executer"
	"github.com/SteakFisher/Redis/app/internal/store"
)

func Pubsub() {
	if store.Pub == nil {
		store.Pub = make(chan string)
	}

	for {
		channelName := <-store.Pub
		channel := store.NameChannel[channelName]

		go func() {
			message := <-channel
			clients := store.ChannelClient[channel]
			for _, v := range clients {
				v.Write(executer.Array(store.StringArr{
					Type: store.Array,
					ArrayVal: []store.StringArr{
						{
							Type:      store.String,
							StringVal: "message",
						},
						{
							Type:      store.String,
							StringVal: channelName,
						},
						{
							Type:      store.String,
							StringVal: message,
						},
					},
				}))

			}
		}()
	}
}
