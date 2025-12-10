package trade

import (
	"dexbe/internal/domains/order"
)

type Trade struct {
	Maker *order.Order
	Taker *order.Order
}

func NewTrade(maker *order.Order, taker *order.Order) *Trade {
	return &Trade{
		Maker: maker,
		Taker: taker,
	}
}
