package router

import (
	"dexbe/internal/infra/api/controllers"

	"github.com/labstack/echo/v4"
)

func RegisterOrderBookRoutes(e *echo.Echo, orderBookController *controller.OrderBookController) {
	orderbooks := e.Group("/orderbook")
	orderbooks.GET("/ws/:in/:out", orderBookController.HandleOrderbookWebSocket)
}
