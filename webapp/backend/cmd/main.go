package main

import (
	"context"
	"dexbe/internal/domains/nonce"
	"dexbe/internal/domains/orderbook"
	"dexbe/internal/domains/registry"
	"dexbe/internal/infra/api"
	"dexbe/internal/infra/api/controllers"
	"dexbe/internal/infra/api/routers"
	"dexbe/internal/infra/eth"
	"dexbe/internal/infra/eth/exchange"
	registryC "dexbe/internal/infra/eth/registry"
	//"dexbe/internal/infra/eth/token"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	godotenv.Load()
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chainId := "31337"
	deployerPrivateKey := "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"
	host := "http://localhost:8545"
	exchangeAddr := os.Getenv("EXCHANGE")
	registryAddr := os.Getenv("TOKENREGISTRY")
	ethClient := eth.GetEthClient(chainId, host, deployerPrivateKey)
	registryContract := registryC.NewRegistryContract(ethClient, registryAddr)
	exchangeContract := exchange.NewExchangeContract(ethClient, exchangeAddr)

	log.Print("Building token registry and pairs...")
	registryStore := registry.NewRegistry(registryContract)
	allTokens := registryContract.GetAllTokens()
	registryStore.Build(&allTokens)
	allSymbols := registryStore.GetAllSymbols()
	orderbs := orderbook.NewOrderBookStore(exchangeContract, allSymbols)
	pollingInterval := 50 * time.Millisecond
	orderbs.StartOracle(ctx, pollingInterval)

	api.StartBroadcast()

	noncer := nonce.NewNonceRegistry()
	convChainId, _ := strconv.Atoi(chainId)
	globalCtrl := controller.NewGlobalController()
	orderCtrl := controller.NewOrderController(orderbs, convChainId, exchangeAddr, registryStore)
	nonceCtrl := controller.NewNonceController(noncer)
	tokenCtrl := controller.NewTokenController(registryContract, orderbs, registryStore)
	orderBookCtrl := controller.NewOrderBookController(orderbs)

	router.RegisterAllRoutes(e, globalCtrl, orderCtrl, orderBookCtrl, nonceCtrl, tokenCtrl)

	log.Println("Starting server on :11223")
	if err := e.Start(":11223"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
