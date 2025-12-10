package controller

import (
	"dexbe/internal/infra/api"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"log"
)

type GlobalController struct {
}

func NewGlobalController() *GlobalController {
	return &GlobalController{}
}

func (ctrl *GlobalController) HandleGlobalWebSocket(ctx echo.Context) error {
	conn, err := api.Upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}
	go func() {
		defer conn.Close()
		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				break
			}
			if messageType == websocket.TextMessage {
				var msg struct {
					Address string `json:"address"`
				}
				if err := json.Unmarshal(data, &msg); err != nil {
					conn.Close()
					log.Println("invalid message:", err)
					continue
				}
				log.Printf("Websocket connected. Address: %+v", msg.Address)
				addr := common.HexToAddress(msg.Address)
				api.AddSubscriber(addr, conn)
			}
		}
	}()

	return nil
}
