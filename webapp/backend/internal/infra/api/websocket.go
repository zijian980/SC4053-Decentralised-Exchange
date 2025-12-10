package api

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
	"net/http"
)

var ( // Hacky solution, but works for project use!
	GlobalCh    chan Message                                = make(chan Message, 100)
	Subscribers map[common.Address]map[*websocket.Conn]bool = map[common.Address]map[*websocket.Conn]bool{}
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	RecipientID common.Address
	Payload     []byte
}

func AddSubscriber(addr common.Address, conn *websocket.Conn) {
	if Subscribers[addr] == nil {
		Subscribers[addr] = make(map[*websocket.Conn]bool)
	}
	Subscribers[addr][conn] = true
}

func RemoveSubscriber(addr common.Address, conn *websocket.Conn) {
	delete(Subscribers[addr], conn)
}

func NotifyUpdate(event string, target common.Address, data any) {
	msg := map[string]any{
		"event": event,
		"data":  data,
	}
	encoded, _ := json.Marshal(msg)
	toSend := Message{
		RecipientID: target,
		Payload:     encoded,
	}

	select {
	case GlobalCh <- toSend:
	default:
	}
}

func StartBroadcast() {
	go func() {
		for msg := range GlobalCh {
			for addr := range Subscribers {
				if msg.RecipientID == addr {
					for conn := range Subscribers[addr] {
						err := conn.WriteMessage(websocket.TextMessage, msg.Payload)
						if err != nil {
							RemoveSubscriber(addr, conn)
							conn.Close()
						}
					}
				}
			}
		}
	}()
}
