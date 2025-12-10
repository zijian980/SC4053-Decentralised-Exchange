package controller

import (
	"dexbe/internal/domains/nonce"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
)

type NonceController struct {
	Registry *nonce.NonceRegistry
}

func NewNonceController(registry *nonce.NonceRegistry) *NonceController {
	return &NonceController{
		Registry: registry,
	}
}

type NonceRequest struct {
	Address string `json:"address"`
}

func (ctrl *NonceController) GetNextNonce(ctx echo.Context) error {
	var req NonceRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}
	hexAddr := common.HexToAddress(req.Address)
	nonce := ctrl.Registry.Inc(hexAddr)
	log.Printf("===INCOMING NONCE INCREMENT===\nAddress: %v\nNew nonce: %v", hexAddr, nonce)
	return ctx.JSON(http.StatusOK, map[string]any{"nonce": nonce})
}
