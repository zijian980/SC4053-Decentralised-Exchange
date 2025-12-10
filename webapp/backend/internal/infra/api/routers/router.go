package router

import (
	"dexbe/internal/infra/api/controllers"
	"github.com/labstack/echo/v4"
)

func RegisterAllRoutes(e *echo.Echo, globalController *controller.GlobalController, orderController *controller.OrderController, orderBookController *controller.OrderBookController, nonceController *controller.NonceController, tokenController *controller.TokenController) {
	RegisterOrderRoutes(e, orderController)
	RegisterOrderBookRoutes(e, orderBookController)
	RegisterNoncewRoutes(e, nonceController)
	RegisterTokenRoutes(e, tokenController)
	e.GET("/ws", globalController.HandleGlobalWebSocket)
}
