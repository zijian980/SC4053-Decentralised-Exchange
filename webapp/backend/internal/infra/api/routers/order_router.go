package router

import (
	"dexbe/internal/infra/api/controllers"
	"github.com/labstack/echo/v4"
)

func RegisterOrderRoutes(e *echo.Echo, orderController *controller.OrderController) {
	orders := e.Group("/order")
	orders.POST("/limit", orderController.SendOrder)
	orders.DELETE("", orderController.CancelOrder)
	orders.GET("/:address", orderController.GetAllOrdersByAddress)
	orders.GET("/history/:address", orderController.GetPastHistory)
	orders.GET("/:in/:out", orderController.GetOrderBookSummary)
}
