package orderbook

import (
	"container/list"
	"math/big"
)

type PriceLevel struct {
	Orders        *list.List
	TotalQuantity *big.Int
}
