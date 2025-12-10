package controller

import (
	"dexbe/internal/domains/order"
	"dexbe/internal/domains/orderbook"
	"dexbe/internal/domains/registry"
	"log"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo/v4"
)

type OrderRequest struct {
	CreatedBy        string           `json:"createdBy"`
	SymbolIn         string           `json:"symbolIn"`
	SymbolOut        string           `json:"symbolOut"`
	AmtIn            string           `json:"amtIn"`
	AmtOut           string           `json:"amtOut"`
	Nonce            string           `json:"nonce"`
	FilledAmtIn      string           `json:"filledAmtIn,omitempty"`
	LimitPrice       string           `json:"limitPrice,omitempty"`
	Status           int              `json:"status,omitempty"`
	ConditionalOrder *SwapInfoRequest `json:"conditionalOrder,omitempty"`
}

type Trigger struct {
	StopPrice string `json:"stopPrice"`
}

type SwapInfoRequest struct {
	Order            *OrderRequest `json:"order"`
	Signature        string        `json:"signature"`
	ConditionTrigger *Trigger      `json:"conditionTriggers"`
}

type GetOrdersRequest struct {
	CreatedBy string `json:"createdBy"`
	SymbolIn  string `json:"symbolIn,omitempty"`
	SymbolOut string `json:"symbolOut,omitempty"`
}

type OrderController struct {
	OrderBookStore  *orderbook.OrderBookStore
	ChainId         int
	ExchangeAddress string
	TokenRegistry   *registry.Registry
}

func NewOrderController(orderBookStore *orderbook.OrderBookStore, chainId int, exchangeAddr string, registry *registry.Registry) *OrderController {
	return &OrderController{
		OrderBookStore:  orderBookStore,
		ChainId:         chainId,
		ExchangeAddress: exchangeAddr,
		TokenRegistry:   registry,
	}
}

func (ctrl *OrderController) SendOrder(ctx echo.Context) error {
	var req SwapInfoRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}
	log.Printf("===INCOMING ORDER===\nORDER: %+v\nORDER SIGNATURE: %+v", req.Order, req.Signature)
	// Order
	convertedOrder := order.NewOrder(req.Order.CreatedBy, req.Order.SymbolIn, req.Order.SymbolOut, req.Order.AmtIn, req.Order.AmtOut, req.Order.Nonce, req.Signature, "", "", req.Order.Status, nil, "")
	decoded_sign, _ := hexutil.Decode(req.Signature)
	verify, err := convertedOrder.VerifyOrder(decoded_sign, big.NewInt(int64(ctrl.ChainId)), common.HexToAddress(ctrl.ExchangeAddress))
	log.Printf("Verify Order: %v", verify)
	if err != nil {
		log.Printf("ERROR: %v", err)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}

	// Conditional Order
	if req.Order.ConditionalOrder != nil {
		conditionalOrderReq := req.Order.ConditionalOrder
		conditionalOrder := conditionalOrderReq.Order
		convertedConditionalOrder := order.NewOrder(conditionalOrder.CreatedBy, conditionalOrder.SymbolIn, conditionalOrder.SymbolOut, conditionalOrder.AmtIn, conditionalOrder.AmtOut, conditionalOrder.Nonce, conditionalOrderReq.Signature, "", "", conditionalOrder.Status, nil, "")
		decoded_sign, _ := hexutil.Decode(req.Signature)
		verify, err := convertedOrder.VerifyOrder(decoded_sign, big.NewInt(int64(ctrl.ChainId)), common.HexToAddress(ctrl.ExchangeAddress))
		log.Printf("Verify Order: %v", verify)
		if err != nil {
			log.Printf("ERROR (Conditionl Order): %v", err)
			return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
		}
		convertedOrder.ConditionalOrder = convertedConditionalOrder
		log.Printf("AHH %+v", req.ConditionTrigger)
		stopPriceBig := new(big.Int)
		stopPriceBig, ok := new(big.Int).SetString(req.ConditionTrigger.StopPrice, 10)
		if !ok {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": "Invalid StopPrice"})
		}
		convertedOrder.ConditionalOrder.TriggerPrice = stopPriceBig
	}
	ctrl.OrderBookStore.AddOrder(convertedOrder)
	return ctx.NoContent(http.StatusOK)
}

func (ctrl *OrderController) GetAllOrdersByAddress(ctx echo.Context) error {
	addr := ctx.Param("address")
	allOrdersByAddress := ctrl.OrderBookStore.GetOrdersByCreator(common.HexToAddress(addr))
	converted := make([]map[string]string, 0, len(allOrdersByAddress))
	for _, o := range allOrdersByAddress {
		converted = append(converted, o.ToStringMap())
	}
	return ctx.JSON(http.StatusOK, map[string]any{"orders": converted})
}

// TODO, cancelled should not remove from map, but rather set status to cancelled and move record to order history
type CancelOrderReq struct {
	CreatedBy  string `json:"createdBy"`
	Nonce      string `json:"nonce"`
	LimitPrice string `json:"limitPrice"`
	SymbolIn   string `json:"symbolIn"`
	SymbolOut  string `json:"symbolOut"`
}

func (ctrl *OrderController) CancelOrder(ctx echo.Context) error {
	var req CancelOrderReq
	if err := ctx.Bind(&req); err != nil {
		log.Printf("ERROR: %+v", err.Error())
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": err.Error()})
	}
	// Change to set status to cancelled (4), using createdby and nonce
	log.Printf("===INCOMING ORDER REMOVAL===\nORDER")
	addr := common.HexToAddress(req.CreatedBy)
	nonce := new(big.Int)
	nonce, ok := nonce.SetString(req.Nonce, 10)
	if !ok {
		log.Println("invalid Nonce:", req.Nonce)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": "Invalid Nonce"})
	}

	limitPrice := new(big.Int)
	limitPrice, ok = limitPrice.SetString(req.LimitPrice, 10)
	if !ok {
		log.Println("invalid LimitPrice:", req.LimitPrice)
		return ctx.JSON(http.StatusBadRequest, map[string]string{"Error": "Invalid LimitPrice"})
	}
	ctrl.OrderBookStore.RemoveOrder(addr, nonce, limitPrice, req.SymbolIn, req.SymbolOut)
	return ctx.NoContent(http.StatusOK)
}

func (ctrl *OrderController) GetOrderBookSummary(ctx echo.Context) error {
	symbolIn := ctx.Param("in")
	symbolOut := ctx.Param("out")
	left, right := orderbook.GetPairKey(symbolIn, symbolOut)
	pairId := left + "/" + right
	reply := ctrl.OrderBookStore.Books[pairId].Snapshot()
	return ctx.JSON(http.StatusOK, reply)
}

func (ctrl *OrderController) GetPastHistory(ctx echo.Context) error {
	address := ctx.Param("address")
	addr := common.HexToAddress(address)
	reply := ctrl.OrderBookStore.GetAllOrderHistoryForAddress(addr)
	return ctx.JSON(http.StatusOK, reply)
}
