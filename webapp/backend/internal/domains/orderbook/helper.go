package orderbook

import (
	"dexbe/internal/domains/order"
	rbtree "github.com/emirpasic/gods/trees/redblacktree"
	"log"
	"math/big"
)

// Helper function to print orders from one side of the book
func printSide(tree *rbtree.Tree, sideName string) {
	if tree.Empty() {
		log.Printf("  %s: (Empty)", sideName)
		return
	}

	log.Printf("--- %s (%d Price Levels) ---", sideName, tree.Size())

	// Use the Iterator to walk through the price levels
	iterator := tree.Iterator()
	iterator.Begin()

	for iterator.Next() {
		// 1. Assert the key as a *big.Int
		priceKey := iterator.Key().(*big.Int)

		// 2. Define the PriceLevel variable
		level := iterator.Value().(*PriceLevel)

		// 3. Convert the *big.Int price key back to a displayable float64
		priceBigFloat := new(big.Float).Quo(
			new(big.Float).SetInt(priceKey),
			new(big.Float).SetInt(PriceFactor),
		)
		priceFloat, _ := priceBigFloat.Float64()

		// Note: Using 18 decimal places for logging to match precision
		log.Printf("  ➡️ Price: %.18f | Aggregated Base Qty: %s", priceFloat, level.TotalQuantity.String())

		// Iterate through the time-priority linked list at this price level
		for e := level.Orders.Front(); e != nil; e = e.Next() {
			order := e.Value.(*order.Order)

			// The AmtIn field consistently holds the remaining quantity for an order on the book.
			// We'll use order.AmtIn for the quantity remaining on the book.
			restingQty := order.AmtIn

			log.Printf("   * Order ID: %s | TotalQty: %s | FilledQty: %s | Price: %.18f",
				order.CreatedBy.String()+order.Nonce.String(), restingQty.String(), order.FilledAmtIn.String(), priceFloat)
		}
	}
}

// PrintOrderBook logs the contents of both the Bids and Asks trees.
func (book *MarketOrderBook) PrintOrderBook() {
	//book.mu.RLock()
	//defer book.mu.RUnlock()

	log.Printf("================= Order Book for %s/%s ==================",
		book.SymbolIn, book.SymbolOut)

	printSide(book.Asks, "Asks")

	log.Printf("---------------------------------------------------------")

	printSide(book.Bids, "Bids")

	log.Printf("=========================================================")
}
