package controller

import (
	"dexbe/internal/domains/orderbook"
	"dexbe/internal/infra/api"
	"encoding/json"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
)

type GetPriceRequest struct {
	SymbolIn  string `json:"symbolIn"`
	SymbolOut string `json:"symbolOut"`
}

type OrderBookController struct {
	OrderBookStore *orderbook.OrderBookStore
}

func NewOrderBookController(store *orderbook.OrderBookStore) *OrderBookController {
	return &OrderBookController{
		OrderBookStore: store,
	}
}

func (ctrl *OrderBookController) HandleOrderbookWebSocket(ctx echo.Context) error {
	in := ctx.Param("in")
	out := ctx.Param("out")
	left, right := orderbook.GetPairKey(in, out)
	pairId := left + "/" + right
	book, ok := ctrl.OrderBookStore.Books[pairId]
	if !ok {
		return ctx.NoContent(http.StatusBadRequest)
	}

	conn, err := api.Upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	}
	book.AddSubscriber(conn)
	snapshot := book.Snapshot()
	data, _ := json.Marshal(snapshot)
	conn.WriteMessage(websocket.TextMessage, data)

	go func() {
		defer func() {
			book.RemoveSubscriber(conn)
			conn.Close()
		}()
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}

		}
	}()

	return nil
}
