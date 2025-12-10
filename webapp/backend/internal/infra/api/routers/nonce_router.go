package router

import (
	"dexbe/internal/infra/api/controllers"
	"github.com/labstack/echo/v4"
)

func RegisterNoncewRoutes(e *echo.Echo, nonceController *controller.NonceController) {
	orders := e.Group("/nonce")
	orders.POST("", nonceController.GetNextNonce)
}
