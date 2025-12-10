package controller

import (
	"dexbe/internal/domains/orderbook"
	"dexbe/internal/domains/registry"
	"dexbe/internal/domains/token"
	registryC "dexbe/internal/infra/eth/registry"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/labstack/echo/v4"
)

type TokenController struct {
	RegistryContract *registryC.RegistryContract
	OrderBookStore   *orderbook.OrderBookStore
	TokenRegistry    *registry.Registry
}

func NewTokenController(contract *registryC.RegistryContract, orderbs *orderbook.OrderBookStore, registry *registry.Registry) *TokenController {
	return &TokenController{RegistryContract: contract, OrderBookStore: orderbs, TokenRegistry: registry}
}

type AddTokenRequest struct {
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
	Address string `json:"address"`
}

func (ctrl *TokenController) AddToken(ctx echo.Context) error {
	var req AddTokenRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}
	ctrl.RegistryContract.AddToken(req.Name, req.Symbol, req.Address)
	addr := common.HexToAddress(req.Address)
	newToken := token.NewToken(req.Name, req.Symbol, addr)
	ctrl.TokenRegistry.Add(newToken)
	allSymbols := ctrl.TokenRegistry.GetAllSymbols()
	for _, symbol := range allSymbols {
		if symbol == newToken.Symbol {
			continue
		}
		ctrl.OrderBookStore.InitializeBook(newToken.Symbol, symbol)
		ctrl.OrderBookStore.InitializeBook(symbol, newToken.Symbol)
	}

	log.Printf("**Token Added**: %s (%s), Order books rebuilt", req.Name, req.Symbol)

	return ctx.NoContent(http.StatusOK)
}
