package router

import (
	"dexbe/internal/infra/api/controllers"

	"github.com/labstack/echo/v4"
)

func RegisterTokenRoutes(e *echo.Echo, tokenController *controller.TokenController) {
	orderbooks := e.Group("/token")
	orderbooks.POST("/add", tokenController.AddToken)
}
