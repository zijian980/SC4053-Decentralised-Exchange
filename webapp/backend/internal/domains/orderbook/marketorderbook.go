package orderbook

import (
	"container/list"
	"context"
	"dexbe/internal/domains/order"
	"dexbe/internal/infra/api"
	"dexbe/internal/infra/eth/exchange"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"sync"

	rbtree "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/gorilla/websocket"
)

type MarketOrderBook struct {
	SymbolIn    string
	SymbolOut   string
	Bids        *rbtree.Tree
	Asks        *rbtree.Tree
	LastPrice   *big.Int
	Mu          sync.RWMutex
	subscribers map[*websocket.Conn]bool
	updateCh    chan []byte
}

const PricePrecision = 18

var PriceFactor = new(big.Int).Exp(big.NewInt(10), big.NewInt(PricePrecision), nil)

func BigIntAscendingComparator(a, b any) int {
	aInt := a.(*big.Int)
	bInt := b.(*big.Int)
	return aInt.Cmp(bInt)
}

func BigIntDescendingComparator(a, b any) int {
	aInt := a.(*big.Int)
	bInt := b.(*big.Int)
	return -aInt.Cmp(bInt)
}

func NewMarketOrderBook(symbolIn, symbolOut string) *MarketOrderBook {
	book := &MarketOrderBook{
		SymbolIn:    symbolIn,
		SymbolOut:   symbolOut,
		Bids:        rbtree.NewWith(BigIntDescendingComparator), // Highest price first
		Asks:        rbtree.NewWith(BigIntAscendingComparator),  // Lowest price first
		LastPrice:   big.NewInt(0),
		subscribers: map[*websocket.Conn]bool{},
		updateCh:    make(chan []byte, 256),
	}
	book.StartBroadcast()
	return book
}

func (book *MarketOrderBook) AddSubscriber(conn *websocket.Conn) {
	book.Mu.Lock()
	book.subscribers[conn] = true
	book.Mu.Unlock()
}

func (book *MarketOrderBook) RemoveSubscriber(conn *websocket.Conn) {
	book.Mu.Lock()
	delete(book.subscribers, conn)
	book.Mu.Unlock()
}

func (book *MarketOrderBook) Snapshot() map[string]any {
	priceFloat := new(big.Float).Quo(
		new(big.Float).SetInt(book.LastPrice),
		new(big.Float).SetInt(PriceFactor),
	)
	lastPrice, _ := priceFloat.Float64()

	return map[string]any{
		"symbolIn":  book.SymbolIn,
		"symbolOut": book.SymbolOut,
		"bids":      book.serializeSide(book.Bids),
		"asks":      book.serializeSide(book.Asks),
		"lastPrice": lastPrice,
	}
}

func (book *MarketOrderBook) StartBroadcast() {
	go func() {
		for msg := range book.updateCh {
			for conn := range book.subscribers {
				err := conn.WriteMessage(websocket.TextMessage, msg)
				if err != nil {
					conn.Close()
					delete(book.subscribers, conn)
				}
			}
		}
	}()
}

func (book *MarketOrderBook) NotifyUpdate(event string, data any) {
	msg := map[string]any{
		"event": event,
		"pair":  fmt.Sprintf("%s/%s", book.SymbolIn, book.SymbolOut),
		"data":  data,
	}
	encoded, _ := json.Marshal(msg)

	select {
	case book.updateCh <- encoded:
	default:
	}
}

func (book *MarketOrderBook) serializeSide(tree *rbtree.Tree) []map[string]any {
	result := []map[string]any{}
	iter := tree.Iterator()
	for iter.Next() {
		priceKey := iter.Key().(*big.Int)
		priceLevel := iter.Value().(*PriceLevel)

		// Convert price to float for readability
		priceFloat := new(big.Float).Quo(
			new(big.Float).SetInt(priceKey),
			new(big.Float).SetInt(PriceFactor),
		)
		price, _ := priceFloat.Float64()

		// Calculate total quantity and count only for orders with status == 0
		totalQuantity := big.NewInt(0)
		activeCount := 0

		for e := priceLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.Status == 0 {
				activeCount++
				remaining := new(big.Int).Sub(o.AmtIn, o.FilledAmtIn)
				totalQuantity.Add(totalQuantity, remaining)
			}
		}

		// Only include price levels with active orders
		if activeCount > 0 {
			// Convert totalQuantity (wei) → token units
			qtyTokens := new(big.Float).Quo(
				new(big.Float).SetInt(totalQuantity),
				new(big.Float).SetInt(PriceFactor),
			)
			qtyFloat, _ := qtyTokens.Float64()

			// If non-zero but tiny, show a minimum of 0.0001
			if qtyFloat > 0 && qtyFloat < 0.0001 {
				qtyFloat = 0.0001
			}

			result = append(result, map[string]any{
				"price":      price,
				"quantity":   totalQuantity.String(),
				"orderCount": activeCount,
			})
		}
	}
	return result
}

// OrderBookStoreInterface defines the interface needed for conditional order management and history tracking
type OrderBookStoreInterface interface {
	AddOrder(*order.Order) error
	StoreConditionalOrder(*order.Order, string) error
	AddToPastHistory(*order.Order)
}

func matchBook(book *MarketOrderBook, exchange *exchange.ExchangeContract, store OrderBookStoreInterface) {
	for {
		bidNode := book.Bids.Left() // Best Bid (Highest Price)
		askNode := book.Asks.Left() // Best Ask (Lowest Price)

		if bidNode == nil || askNode == nil {
			break // One or both sides empty
		}

		bidPriceKey := bidNode.Key.(*big.Int)
		bidLevel := bidNode.Value.(*PriceLevel)

		askPriceKey := askNode.Key.(*big.Int)
		askLevel := askNode.Value.(*PriceLevel)

		// Orders match when bid price >= ask price
		if bidPriceKey.Cmp(askPriceKey) < 0 {
			break
		}

		// Check if any order at these price levels is pending
		// If so, exit matchBook to prevent double-matching
		hasPending := false
		for e := bidLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.Status == 1 { // Pending transaction
				hasPending = true
				break
			}
		}
		if hasPending {
			log.Printf("Bid level has pending orders, waiting for confirmation")
			return
		}

		for e := askLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.Status == 1 { // Pending transaction
				hasPending = true
				break
			}
		}
		if hasPending {
			log.Printf("Ask level has pending orders, waiting for confirmation")
			return
		}

		// Get the oldest ready order from each level (FIFO)
		var bidElem *list.Element
		var bidOrder *order.Order
		for e := bidLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.Status == 0 { // Ready to match (not pending, not filled)
				bidElem = e
				bidOrder = o
				break
			}
		}

		var askElem *list.Element
		var askOrder *order.Order
		for e := askLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.Status == 0 { // Ready to match (not pending, not filled)
				askElem = e
				askOrder = o
				break
			}
		}

		// No available orders to match at these price levels
		if bidElem == nil || askElem == nil {
			log.Printf("No ready orders at price levels - Bid: %v, Ask: %v", bidPriceKey, askPriceKey)
			break
		}

		// Prevent self-matching
		if bidOrder.CreatedBy.Cmp(askOrder.CreatedBy) == 0 {
			// log.Printf("**Self-Match Prevention**: Skipping match - same user (%s) on both sides",
			// bidOrder.CreatedBy.Hex()[:10])
			break
		}

		// Initialize FilledAmtIn if nil (for new orders)
		if bidOrder.FilledAmtIn == nil {
			bidOrder.FilledAmtIn = big.NewInt(0)
		}
		if askOrder.FilledAmtIn == nil {
			askOrder.FilledAmtIn = big.NewInt(0)
		}

		// Calculate REMAINING amounts
		// Bid: AmtIn = quote currency (what they're spending), AmtOut = base currency (what they want)
		// Ask: AmtIn = base currency (what they're selling), AmtOut = quote currency (what they want)
		bidRemainingIn := new(big.Int).Sub(bidOrder.AmtIn, bidOrder.FilledAmtIn) // Quote remaining to spend
		askRemainingIn := new(big.Int).Sub(askOrder.AmtIn, askOrder.FilledAmtIn) // Base remaining to sell

		// Remove fully filled orders BEFORE attempting to match
		if bidRemainingIn.Cmp(big.NewInt(0)) <= 0 {
			bidLevel.Orders.Remove(bidElem)
			bidOrder.Status = 2
			log.Printf("**Order Fully Filled**: Bid %s/%s removed (filled %s/%s)",
				bidOrder.CreatedBy.Hex()[:10], bidOrder.Nonce.String(),
				bidOrder.FilledAmtIn.String(), bidOrder.AmtIn.String())
			if bidLevel.Orders.Len() == 0 {
				book.Bids.Remove(bidPriceKey)
			}
			continue
		}
		if askRemainingIn.Cmp(big.NewInt(0)) <= 0 {
			askLevel.Orders.Remove(askElem)
			askOrder.Status = 2
			log.Printf("**Order Fully Filled**: Ask %s/%s removed (filled %s/%s)",
				askOrder.CreatedBy.Hex()[:10], askOrder.Nonce.String(),
				askOrder.FilledAmtIn.String(), askOrder.AmtIn.String())
			if askLevel.Orders.Len() == 0 {
				book.Asks.Remove(askPriceKey)
			}
			continue
		}

		// Calculate how much base currency the bid wants to receive (proportional to remaining quote)
		bidRemainingOut := new(big.Int).Mul(bidOrder.AmtOut, bidRemainingIn)
		bidRemainingOut.Div(bidRemainingOut, bidOrder.AmtIn)

		// Calculate how much quote currency the ask wants to receive (proportional to remaining base)
		askRemainingOut := new(big.Int).Mul(askOrder.AmtOut, askRemainingIn)
		askRemainingOut.Div(askRemainingOut, askOrder.AmtIn)

		// Determine the maximum tradeable base quantity
		// This is limited by:
		// 1. How much base the ask is selling (askRemainingIn)
		// 2. How much base the bid wants to buy (bidRemainingOut)
		tradeBaseQty := new(big.Int)
		if askRemainingIn.Cmp(bidRemainingOut) < 0 {
			// Ask has less base to sell than bid wants
			tradeBaseQty.Set(askRemainingIn)
		} else {
			// Bid wants less base than ask is selling
			tradeBaseQty.Set(bidRemainingOut)
		}

		if tradeBaseQty.Cmp(big.NewInt(0)) <= 0 {
			log.Printf("**Warning**: Invalid trade quantity, skipping match")
			break
		}

		// Calculate execution price and quote quantity
		// Use the ask price (maker price) as execution price
		executionPrice := askPriceKey
		tradePriceFloat := new(big.Float).Quo(
			new(big.Float).SetInt(executionPrice),
			new(big.Float).SetInt(PriceFactor),
		)
		priceFloat64, _ := tradePriceFloat.Float64()

		// Calculate quote amount needed using INTEGER math: (tradeBaseQty * executionPrice) / PriceFactor
		tradeQuoteQty := new(big.Int).Mul(tradeBaseQty, executionPrice)
		tradeQuoteQty.Div(tradeQuoteQty, PriceFactor)

		// Ensure the bid has enough quote remaining
		if tradeQuoteQty.Cmp(bidRemainingIn) > 0 {
			// Bid doesn't have enough quote, recalculate base amount using INTEGER math
			tradeQuoteQty.Set(bidRemainingIn)
			// tradeBaseQty = (tradeQuoteQty * PriceFactor) / executionPrice
			tradeBaseQty = new(big.Int).Mul(tradeQuoteQty, PriceFactor)
			tradeBaseQty.Div(tradeBaseQty, executionPrice)

			// Ensure we don't try to trade more base than ask has
			if tradeBaseQty.Cmp(askRemainingIn) > 0 {
				tradeBaseQty.Set(askRemainingIn)
				// Recalculate quote to match using INTEGER math
				tradeQuoteQty = new(big.Int).Mul(tradeBaseQty, executionPrice)
				tradeQuoteQty.Div(tradeQuoteQty, PriceFactor)
			}
		}

		// Dust handling: If this would complete either order within a tiny margin, use exact remaining
		dustThreshold := big.NewInt(100) // 100 wei tolerance

		bidAfterTrade := new(big.Int).Sub(bidRemainingIn, tradeQuoteQty)
		askAfterTrade := new(big.Int).Sub(askRemainingIn, tradeBaseQty)

		if bidAfterTrade.Cmp(dustThreshold) <= 0 && bidAfterTrade.Cmp(big.NewInt(0)) > 0 {
			// Bid has dust remaining, consume it all
			log.Printf("Bid dust detected (%s wei), consuming full remaining amount", bidAfterTrade.String())
			tradeQuoteQty.Set(bidRemainingIn)
			// Recalculate base amount
			tradeBaseQty = new(big.Int).Mul(tradeQuoteQty, PriceFactor)
			tradeBaseQty.Div(tradeBaseQty, executionPrice)
			// Ensure we don't exceed ask's remaining
			if tradeBaseQty.Cmp(askRemainingIn) > 0 {
				tradeBaseQty.Set(askRemainingIn)
			}
		}

		if askAfterTrade.Cmp(dustThreshold) <= 0 && askAfterTrade.Cmp(big.NewInt(0)) > 0 {
			// Ask has dust remaining, consume it all
			log.Printf("Ask dust detected (%s wei), consuming full remaining amount", askAfterTrade.String())
			tradeBaseQty.Set(askRemainingIn)
			// Recalculate quote amount
			tradeQuoteQty = new(big.Int).Mul(tradeBaseQty, executionPrice)
			tradeQuoteQty.Div(tradeQuoteQty, PriceFactor)
			// Ensure we don't exceed bid's remaining
			if tradeQuoteQty.Cmp(bidRemainingIn) > 0 {
				tradeQuoteQty.Set(bidRemainingIn)
			}
		}

		if tradeBaseQty.Cmp(big.NewInt(0)) <= 0 || tradeQuoteQty.Cmp(big.NewInt(0)) <= 0 {
			log.Printf("**Warning**: Invalid trade quantities after calculation, skipping match")
			break
		}

		log.Printf("**Trade**: %s/%s | Buyer: %s | Seller: %s | Base: %s | Quote: %s | Price: %.6f",
			book.SymbolIn, book.SymbolOut,
			bidOrder.CreatedBy.Hex()[:10],
			askOrder.CreatedBy.Hex()[:10],
			tradeBaseQty.String(),
			tradeQuoteQty.String(),
			priceFloat64,
		)

		log.Printf("  Before: Bid %s/%s (%.1f%%) | Ask %s/%s (%.1f%%)",
			bidOrder.FilledAmtIn.String(), bidOrder.AmtIn.String(),
			percent(bidOrder.FilledAmtIn, bidOrder.AmtIn),
			askOrder.FilledAmtIn.String(), askOrder.AmtIn.String(),
			percent(askOrder.FilledAmtIn, askOrder.AmtIn),
		)
		// Mark orders as PendingConfirmation before sending transaction
		bidOrder.Status = 1
		askOrder.Status = 1

		log.Printf("===SENDING TRANSACTION TO BLOCKCHAIN===")
		log.Printf("ASK: %s/%s | BID: %s/%s | Fill: %s",
			askOrder.CreatedBy.Hex()[:10], askOrder.Nonce.String(),
			bidOrder.CreatedBy.Hex()[:10], bidOrder.Nonce.String(),
			tradeBaseQty.String())

		// Use ExecuteMatch to submit the transaction
		tx, err := exchange.ExecuteMatch(askOrder, bidOrder, tradeBaseQty)

		if err != nil {
			log.Printf("ERROR EXECUTING MATCH: %+v", err)
			bidOrder.Status = 0
			askOrder.Status = 0
			continue // Try next match
		}
		api.NotifyUpdate("TransactionChange", askOrder.CreatedBy, askOrder.ToStringMap())
		api.NotifyUpdate("TransactionChange", bidOrder.CreatedBy, bidOrder.ToStringMap())

		txHash := tx.Hash().Hex()
		log.Printf("Transaction sent: %s", txHash)

		// Create copies of values needed in goroutine to avoid race conditions
		finalBidOrder := bidOrder
		finalAskOrder := askOrder
		finalBidElem := bidElem
		finalAskElem := askElem
		finalBidLevel := bidLevel
		finalAskLevel := askLevel
		finalBidPriceKey := new(big.Int).Set(bidPriceKey)
		finalAskPriceKey := new(big.Int).Set(askPriceKey)
		finalTradeBaseQty := new(big.Int).Set(tradeBaseQty)
		finalTradeQuoteQty := new(big.Int).Set(tradeQuoteQty)
		finalExecutionPrice := new(big.Int).Set(executionPrice)

		// Goroutine to wait for confirmation
		go func() {
			// Wait for transaction receipt
			minedStatus := make(chan int)
			go exchange.Client.CheckTxReceipt(context.Background(), tx, minedStatus)

			result := <-minedStatus

			book.Mu.Lock()
			defer book.Mu.Unlock()

			if result == 1 { // Success
				log.Printf(" Transaction %s confirmed", txHash)

				// Update last price using the execution price (ask price)
				book.LastPrice.Set(finalExecutionPrice)

				priceFloat := new(big.Float).Quo(
					new(big.Float).SetInt(book.LastPrice),
					new(big.Float).SetInt(PriceFactor),
				)
				lastPriceFloat64, _ := priceFloat.Float64()
				log.Printf("📊 Last Price Updated: %.6f for %s/%s",
					lastPriceFloat64, book.SymbolIn, book.SymbolOut)

				// Add transaction hash to both orders
				finalBidOrder.TransactionHashes = append(finalBidOrder.TransactionHashes, txHash)
				finalAskOrder.TransactionHashes = append(finalAskOrder.TransactionHashes, txHash)

				// Update filled amounts (cumulative)
				// Bid fills with quote currency, Ask fills with base currency
				finalBidOrder.FilledAmtIn.Add(finalBidOrder.FilledAmtIn, finalTradeQuoteQty)
				finalAskOrder.FilledAmtIn.Add(finalAskOrder.FilledAmtIn, finalTradeBaseQty)

				log.Printf("  After: Bid %s/%s (%.1f%%) | Ask %s/%s (%.1f%%)",
					finalBidOrder.FilledAmtIn.String(), finalBidOrder.AmtIn.String(),
					percent(finalBidOrder.FilledAmtIn, finalBidOrder.AmtIn),
					finalAskOrder.FilledAmtIn.String(), finalAskOrder.AmtIn.String(),
					percent(finalAskOrder.FilledAmtIn, finalAskOrder.AmtIn),
				)

				// Update price level quantities
				finalBidLevel.TotalQuantity.Sub(finalBidLevel.TotalQuantity, finalTradeQuoteQty)
				finalAskLevel.TotalQuantity.Sub(finalAskLevel.TotalQuantity, finalTradeBaseQty)

				bidNewRemaining := new(big.Int).Sub(finalBidOrder.AmtIn, finalBidOrder.FilledAmtIn)
				askNewRemaining := new(big.Int).Sub(finalAskOrder.AmtIn, finalAskOrder.FilledAmtIn)

				// Handle bid order completion
				if bidNewRemaining.Cmp(big.NewInt(0)) == 0 {
					// Fully filled - status 3, add to history, then set to 2
					finalBidOrder.Status = 3
					store.AddToPastHistory(finalBidOrder)

					finalBidLevel.Orders.Remove(finalBidElem)
					finalBidOrder.Status = 2
					log.Printf("**Order Fully Filled**: Bid %s/%s (100%%) - Added to history",
						finalBidOrder.CreatedBy.Hex()[:10], finalBidOrder.Nonce.String())

					book.NotifyUpdate("orderbook_update", book.Snapshot())

					if finalBidOrder.ConditionalOrder != nil {
						log.Printf("**Conditional Order Detected**: Storing conditional order for Bid %s/%s",
							finalBidOrder.CreatedBy.Hex()[:10], finalBidOrder.Nonce.String())
						parentOrderID := fmt.Sprintf("%s-%s", finalBidOrder.CreatedBy.Hex(), finalBidOrder.Nonce.String())
						conditionalOrderCopy := finalBidOrder.ConditionalOrder
						parentIDCopy := parentOrderID
						go func() {
							err := store.StoreConditionalOrder(conditionalOrderCopy, parentIDCopy)
							if err != nil {
								log.Printf("**Error Storing Conditional Order**: %v", err)
							}
						}()
					}
				} else {
					// Partially filled - status 5, add to history, then set back to 0
					finalBidOrder.Status = 5
					store.AddToPastHistory(finalBidOrder)

					finalBidOrder.Status = 0
					log.Printf("**Order Partially Filled**: Bid %s/%s (%.1f%%) - Added to history",
						finalBidOrder.CreatedBy.Hex()[:10], finalBidOrder.Nonce.String(),
						percent(finalBidOrder.FilledAmtIn, finalBidOrder.AmtIn))
				}

				// Handle ask order completion
				if askNewRemaining.Cmp(big.NewInt(0)) == 0 {
					// Fully filled - status 3, add to history, then set to 2
					finalAskOrder.Status = 3
					store.AddToPastHistory(finalAskOrder)

					finalAskLevel.Orders.Remove(finalAskElem)
					finalAskOrder.Status = 2
					log.Printf("**Order Fully Filled**: Ask %s/%s (100%%) - Added to history",
						finalAskOrder.CreatedBy.Hex()[:10], finalAskOrder.Nonce.String())

					book.NotifyUpdate("orderbook_update", book.Snapshot())

					if finalAskOrder.ConditionalOrder != nil {
						log.Printf("**Conditional Order Detected**: Storing conditional order for Ask %s/%s",
							finalAskOrder.CreatedBy.Hex()[:10], finalAskOrder.Nonce.String())
						parentOrderID := fmt.Sprintf("%s-%s", finalAskOrder.CreatedBy.Hex(), finalAskOrder.Nonce.String())
						conditionalOrderCopy := finalAskOrder.ConditionalOrder
						parentIDCopy := parentOrderID
						go func() {
							err := store.StoreConditionalOrder(conditionalOrderCopy, parentIDCopy)
							if err != nil {
								log.Printf("**Error Storing Conditional Order**: %v", err)
							}
						}()
					}
				} else {
					// Partially filled - status 5, add to history, then set back to 0
					finalAskOrder.Status = 5
					store.AddToPastHistory(finalAskOrder)

					finalAskOrder.Status = 0
					log.Printf("**Order Partially Filled**: Ask %s/%s (%.1f%%) - Added to history",
						finalAskOrder.CreatedBy.Hex()[:10], finalAskOrder.Nonce.String(),
						percent(finalAskOrder.FilledAmtIn, finalAskOrder.AmtIn))
				}

				if finalBidLevel.Orders.Len() == 0 {
					book.Bids.Remove(finalBidPriceKey)
					book.NotifyUpdate("orderbook_update", book.Snapshot())
				}
				if finalAskLevel.Orders.Len() == 0 {
					book.Asks.Remove(finalAskPriceKey)
					book.NotifyUpdate("orderbook_update", book.Snapshot())
				}

				api.NotifyUpdate("TransactionChange", finalAskOrder.CreatedBy, finalAskOrder.ToStringMap())
				api.NotifyUpdate("TransactionChange", finalBidOrder.CreatedBy, finalBidOrder.ToStringMap())

				// Check conditional trigger based on new last price**
				go store.(*OrderBookStore).ConditionalOrderStore.CheckPriceTriggersForBook(
					book.SymbolIn,
					book.SymbolOut,
					book.LastPrice,
				)

			} else {
				log.Printf("❌ Transaction %s failed or reverted", txHash)
				finalBidOrder.Status = 0
				finalAskOrder.Status = 0

				api.NotifyUpdate("TransactionChange", finalAskOrder.CreatedBy, finalAskOrder.ToStringMap())
				api.NotifyUpdate("TransactionChange", finalBidOrder.CreatedBy, finalBidOrder.ToStringMap())
			}
		}()

		// Exit after submitting ONE transaction to prevent double-matching
		// The matching engine will be called again after the transaction confirms
		return
	}
}

func percent(filled, total *big.Int) float64 {
	if total.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	f := new(big.Float).SetInt(filled)
	t := new(big.Float).SetInt(total)
	p, _ := new(big.Float).Quo(f, t).Float64()
	return p * 100
}
