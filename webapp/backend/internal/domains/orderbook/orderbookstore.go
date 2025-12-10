package orderbook

import (
	"container/list"
	"context"
	"dexbe/internal/domains/order"
	"dexbe/internal/infra/api"
	"dexbe/internal/infra/eth/exchange"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	rbtree "github.com/emirpasic/gods/trees/redblacktree"
	"github.com/ethereum/go-ethereum/common"
)

// UnmatchedOrder represents an order that couldn't be fully matched in its book
type UnmatchedOrder struct {
	Order        *order.Order
	RemainingIn  *big.Int
	RemainingOut *big.Int
	Book         *MarketOrderBook
	PriceLevel   *PriceLevel
	Element      *list.Element
}

// RingPath represents a circular trade path across multiple order books
type RingPath struct {
	Orders       []*order.Order
	Books        []*MarketOrderBook
	PriceLevels  []*PriceLevel
	Elements     []*list.Element
	TradeAmounts []*big.Int
}

type OrderBookStore struct {
	Books                 map[string]*MarketOrderBook
	Exchange              *exchange.ExchangeContract
	mu                    sync.RWMutex
	OrderHistory          map[string]order.Order
	PastTransactions      map[string][]order.Order
	maxRingDepth          int
	ringMatchingEnabled   bool
	ConditionalOrderStore *ConditionalOrderStore
	PastHistoryStore      map[common.Address]map[string][]order.Order
}

type MarketPrice struct {
	BestBid   *big.Int
	BestAsk   *big.Int
	MidPrice  *big.Int
	Spread    *big.Int
	LastPrice *big.Int
}

func NewOrderBookStore(exchange *exchange.ExchangeContract, symbols []string) *OrderBookStore {
	store := &OrderBookStore{
		Books:               make(map[string]*MarketOrderBook),
		Exchange:            exchange,
		maxRingDepth:        5,    // default max depth of 5
		ringMatchingEnabled: true, // enable by default
		PastHistoryStore:    make(map[common.Address]map[string][]order.Order),
	}

	// Initialize conditional order store
	store.ConditionalOrderStore = NewConditionalOrderStore(store)

	for i := 0; i < len(symbols); i++ {
		for j := i + 1; j < len(symbols); j++ {
			base := symbols[i]
			quote := symbols[j]

			left, right := GetPairKey(base, quote)
			pairKey := left + "/" + right
			store.Books[pairKey] = NewMarketOrderBook(left, right)
		}
	}

	log.Printf("OrderBookStore initialized with %d books (ring matching: enabled, max depth: %d)",
		len(store.Books), store.maxRingDepth)
	return store
}

// AddToPastHistory adds a copy of an order to the PastHistoryStore
func (store *OrderBookStore) AddToPastHistory(o *order.Order) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Create a DEEP copy of the order
	orderCopy := o.DeepCopy()

	// Initialize the address map if it doesn't exist
	if store.PastHistoryStore[o.CreatedBy] == nil {
		store.PastHistoryStore[o.CreatedBy] = make(map[string][]order.Order)
	}

	nonceKey := o.Nonce.String()
	// Append to the history slice for this nonce
	store.PastHistoryStore[o.CreatedBy][nonceKey] = append(
		store.PastHistoryStore[o.CreatedBy][nonceKey],
		*orderCopy, // Dereference the deep copy
	)

	log.Printf("**History**: Added order snapshot to PastHistoryStore - Address: %s, Nonce: %s, Status: %d",
		o.CreatedBy.Hex()[:10], nonceKey, o.Status)
}

// GetOrderHistory retrieves all historical snapshots for a specific order
func (store *OrderBookStore) GetOrderHistory(creator common.Address, nonce *big.Int) []order.Order {
	store.mu.RLock()
	defer store.mu.RUnlock()

	if addressMap, exists := store.PastHistoryStore[creator]; exists {
		if history, exists := addressMap[nonce.String()]; exists {
			return history
		}
	}
	return []order.Order{}
}

// GetAllOrderHistoryForAddress retrieves all order histories for a specific address
func (store *OrderBookStore) GetAllOrderHistoryForAddress(creator common.Address) map[string][]order.Order {
	store.mu.RLock()
	defer store.mu.RUnlock()
	if addressMap, exists := store.PastHistoryStore[creator]; exists {
		return addressMap
	}
	return make(map[string][]order.Order)
}

// SetRingMatchingEnabled enables or disables ring matching
func (store *OrderBookStore) SetRingMatchingEnabled(enabled bool) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.ringMatchingEnabled = enabled
	log.Printf("Ring matching %s", map[bool]string{true: "enabled", false: "disabled"}[enabled])
}

// SetMaxRingDepth sets the maximum ring depth for ring matching
func (store *OrderBookStore) SetMaxRingDepth(depth int) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.maxRingDepth = depth
	log.Printf("Max ring depth set to %d", depth)
}

// GetPairKey returns base and quote tokens in canonical order
// The FIRST token returned is always the BASE, second is QUOTE
func GetPairKey(tokenA, tokenB string) (base, quote string) {
	if tokenA < tokenB {
		return tokenA, tokenB
	}
	return tokenB, tokenA
}

func (store *OrderBookStore) StartOracle(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Println("**Oracle Started**: Matching engine running")
		for {
			select {
			case <-ctx.Done():
				log.Println("**Oracle Stopped**: Matching engine shut down")
				return
			case <-ticker.C:
				store.matchAllBooks()
			}
		}
	}()
}

func (store *OrderBookStore) matchAllBooks() {
	store.mu.RLock()
	books := make([]*MarketOrderBook, 0, len(store.Books))
	for _, book := range store.Books {
		books = append(books, book)
	}
	ringEnabled := store.ringMatchingEnabled
	store.mu.RUnlock()

	if ringEnabled {
		// log.Printf(" Starting multi-book matching cycle (ring matching enabled)...")

		// Phase 1: Direct matching within each book
		for _, book := range books {
			book.Mu.Lock()
			matchBook(book, store.Exchange, store)
			book.Mu.Unlock()
		}

		// Phase 2: Ring matching - iterate until no more rings found
		ringRound := 0
		for {
			ringRound++
			// log.Printf(" Ring matching round %d", ringRound)

			// Collect unmatched orders each round
			allUnmatched := []*UnmatchedOrder{}
			for _, book := range books {
				book.Mu.Lock()
				unmatched := store.collectUnmatchedOrders(book)
				book.Mu.Unlock()
				allUnmatched = append(allUnmatched, unmatched...)
			}

			// log.Printf(" Found %d unmatched orders in round %d", len(allUnmatched), ringRound)

			if len(allUnmatched) == 0 {
				// log.Printf(" No unmatched orders remaining")
				break
			}

			// Try to find and execute ONE ring
			ringsFound := store.findAndExecuteOneRing(allUnmatched)

			if ringsFound == 0 {
				//log.Printf(" No more rings found in round %d", ringRound)
				break
			}

			// If we found a ring, loop again to find more with updated amounts
			// Limit to prevent infinite loops
			if ringRound >= 10 {
				log.Printf("  Ring matching round limit reached (%d rounds)", ringRound)
				break
			}
		}
	} else {
		// Legacy direct matching only
		for _, book := range books {
			book.Mu.Lock()
			matchBook(book, store.Exchange, store)
			book.Mu.Unlock()
		}
	}
}

// collectUnmatchedOrders collects orders with remaining unfilled quantity
func (store *OrderBookStore) collectUnmatchedOrders(book *MarketOrderBook) []*UnmatchedOrder {
	unmatched := []*UnmatchedOrder{}

	// Helper to collect from a tree
	collectFromTree := func(tree *rbtree.Tree) {
		iter := tree.Iterator()
		for iter.Next() {
			priceLevel := iter.Value().(*PriceLevel)
			for e := priceLevel.Orders.Front(); e != nil; e = e.Next() {
				order := e.Value.(*order.Order)
				if order.Status != 0 {
					continue // Skip pending or filled orders
				}

				if order.FilledAmtIn == nil {
					order.FilledAmtIn = big.NewInt(0)
				}

				remainingIn := new(big.Int).Sub(order.AmtIn, order.FilledAmtIn)
				if remainingIn.Cmp(big.NewInt(0)) > 0 {
					remainingOut := new(big.Int).Mul(order.AmtOut, remainingIn)
					remainingOut.Div(remainingOut, order.AmtIn)

					unmatched = append(unmatched, &UnmatchedOrder{
						Order:        order,
						RemainingIn:  remainingIn,
						RemainingOut: remainingOut,
						Book:         book,
						PriceLevel:   priceLevel,
						Element:      e,
					})
				}
			}
		}
	}

	// Collect from both sides
	collectFromTree(book.Bids)
	collectFromTree(book.Asks)

	return unmatched
}

// findAndExecuteOneRing finds and executes ONE ring, then returns
func (store *OrderBookStore) findAndExecuteOneRing(unmatched []*UnmatchedOrder) int {
	//log.Printf(" Searching for a ring trade among %d orders...", len(unmatched))

	for _, unmatchedOrder := range unmatched {
		//orderKey := getOrderKey(unmatchedOrder.Order)

		// Skip if order has no remaining amount (might have been filled in previous ring this round)
		if unmatchedOrder.RemainingIn.Cmp(big.NewInt(0)) <= 0 {
			//log.Printf("    Skipping order %s (no remaining amount)", orderKey[:20])
			continue
		}

		// Try to find a ring starting from this order
		ring := store.findRing(unmatchedOrder)

		if ring != nil {
			//log.Printf(" Found ring with %d orders", len(ring.Orders))

			// Execute the ring
			err := store.executeRing(ring)
			if err != nil {
				//log.Printf(" Ring execution failed: %v", err)
				continue // Try to find another ring
			}

			log.Printf(" Ring executed successfully")
			return 1 // Found and executed one ring
		}
	}

	return 0
}

func (store *OrderBookStore) findRing(start *UnmatchedOrder) *RingPath {

	targetToken := start.Order.SymbolIn   // Token we want to get back to (what we're giving)
	currentToken := start.Order.SymbolOut // Token we currently have (what we want)

	//log.Printf(" Searching for ring: Start order gives %s wants %s | Target: get back %s | Current: %s",
	//	start.Order.SymbolIn, start.Order.SymbolOut, targetToken, currentToken)
	//log.Printf("   Starting order remaining: %s %s", start.RemainingIn.String(), start.Order.SymbolIn)

	initialPath := &RingPath{
		Orders:       []*order.Order{start.Order},
		Books:        []*MarketOrderBook{start.Book},
		PriceLevels:  []*PriceLevel{start.PriceLevel},
		Elements:     []*list.Element{start.Element},
		TradeAmounts: []*big.Int{start.RemainingIn},
	}

	visited := make(map[string]bool)
	visited[getOrderKey(start.Order)] = true

	// Track users in the ring to prevent self-matching
	usersInRing := make(map[common.Address]bool)
	usersInRing[start.Order.CreatedBy] = true

	result := store.dfsRing(currentToken, targetToken, initialPath, visited, usersInRing, 1)

	if result != nil {
		log.Printf(" Ring found: %s", store.ringToString(result))
		log.Printf("   Ring orders and capacities:")
		for i, order := range result.Orders {
			log.Printf("     %d. %s gives %s wants %s, capacity: %s",
				i+1, getOrderKey(order)[:20], order.SymbolIn, order.SymbolOut,
				result.TradeAmounts[i].String())
		}
	}

	return result
}

func (store *OrderBookStore) dfsRing(
	currentToken, targetToken string,
	path *RingPath,
	visited map[string]bool,
	usersInRing map[common.Address]bool,
	depth int,
) *RingPath {
	// Depth limit
	if depth > store.maxRingDepth {
		return nil
	}

	//log.Printf("  [DFS depth=%d] currentToken=%s, targetToken=%s, pathLen=%d",
	//	depth, currentToken, targetToken, len(path.Orders))

	// Check if we've completed the ring
	if currentToken == targetToken && depth > 1 {
		//log.Printf("  [DFS] Ring closed! Validating...")
		// Validate the ring is economically viable
		if store.validateRing(path) {
			return path
		}
		return nil
	}

	// Find orders where someone is GIVING currentToken (their SymbolIn = currentToken)
	//log.Printf("  [DFS] Looking for orders where SymbolIn=%s (giving %s)", currentToken, currentToken)
	ordersFound := 0

	for _, book := range store.Books {
		book.Mu.RLock()

		// Check both bid and ask sides for orders giving currentToken
		checkOrders := func(tree *rbtree.Tree, side string) *RingPath {
			if tree == nil {
				return nil
			}

			iter := tree.Iterator()
			for iter.Next() {
				priceLevel := iter.Value().(*PriceLevel)

				for e := priceLevel.Orders.Front(); e != nil; e = e.Next() {
					askOrder := e.Value.(*order.Order)

					// We're looking for orders that give currentToken (SymbolIn = currentToken)
					if askOrder.SymbolIn != currentToken {
						continue
					}

					ordersFound++
					//log.Printf("  [DFS] Found order in %s %s: %s gives %s wants %s",
					//	bookKey, side, getOrderKey(askOrder)[:20], askOrder.SymbolIn, askOrder.SymbolOut)

					// Skip if not ready
					if askOrder.Status != 0 {
						// log.Printf("  [DFS] Order %s skipped (status=%d)", getOrderKey(askOrder), askOrder.Status)
						continue
					}

					// NEW: Prevent self-matching in rings
					if usersInRing[askOrder.CreatedBy] {
						// log.Printf("  [DFS] Order %s skipped (user %s already in ring)",
						// 	getOrderKey(askOrder)[:20], askOrder.CreatedBy.Hex()[:10])
						continue
					}

					orderKey := getOrderKey(askOrder)
					if visited[orderKey] {
						// log.Printf("  [DFS] Order %s already visited", orderKey)
						continue
					}

					// Initialize filled amount
					if askOrder.FilledAmtIn == nil {
						askOrder.FilledAmtIn = big.NewInt(0)
					}

					remainingIn := new(big.Int).Sub(askOrder.AmtIn, askOrder.FilledAmtIn)
					if remainingIn.Cmp(big.NewInt(0)) <= 0 {
						// log.Printf("  [DFS] Order %s has no remaining amount", orderKey)
						continue
					}

					// log.Printf("  [DFS]  Using order: %s gives %s wants %s (remaining: %s)",
					// 	orderKey[:20], askOrder.SymbolIn, askOrder.SymbolOut, remainingIn.String())

					// Create new path with this order
					newPath := &RingPath{
						Orders:       append(append([]*order.Order{}, path.Orders...), askOrder),
						Books:        append(append([]*MarketOrderBook{}, path.Books...), book),
						PriceLevels:  append(append([]*PriceLevel{}, path.PriceLevels...), priceLevel),
						Elements:     append(append([]*list.Element{}, path.Elements...), e),
						TradeAmounts: append(append([]*big.Int{}, path.TradeAmounts...), remainingIn),
					}

					visited[orderKey] = true

					// NEW: Add user to ring tracking
					newUsersInRing := make(map[common.Address]bool)
					for user := range usersInRing {
						newUsersInRing[user] = true
					}
					newUsersInRing[askOrder.CreatedBy] = true

					result := store.dfsRing(
						askOrder.SymbolOut, // Next token we have (what this order wants)
						targetToken,
						newPath,
						visited,
						newUsersInRing, // Pass updated user map
						depth+1,
					)

					if result != nil {
						return result
					}

					delete(visited, orderKey) // Backtrack
				}
			}
			return nil
		}

		// Check asks
		result := checkOrders(book.Asks, "ASK")
		if result != nil {
			book.Mu.RUnlock()
			return result
		}

		// Check bids
		result = checkOrders(book.Bids, "BID")
		if result != nil {
			book.Mu.RUnlock()
			return result
		}

		book.Mu.RUnlock()
	}

	if ordersFound == 0 {
		// log.Printf("  [DFS] No orders found giving %s", currentToken)
	} else {
		// log.Printf("  [DFS] Found %d orders giving %s but none usable", ordersFound, currentToken)
	}

	return nil
}

// ============================================================================
// RING VALIDATION AND EXECUTION
// ============================================================================

func (store *OrderBookStore) validateRing(ring *RingPath) bool {
	// Check that the ring is closed (last order's output matches first order's input)
	if len(ring.Orders) < 2 {
		return false
	}

	firstOrder := ring.Orders[0]
	lastOrder := ring.Orders[len(ring.Orders)-1]

	// The ring is closed if:
	// - First order gives token X (SymbolIn)
	// - Last order wants token X (SymbolOut)
	if lastOrder.SymbolOut != firstOrder.SymbolIn {
		// log.Printf(" Ring not closed: last order wants %s but first order gives %s",
		// lastOrder.SymbolOut, firstOrder.SymbolIn)
		return false
	}

	// CRITICAL: Prevent rings within the same order book
	// A 2-order ring where both orders are in the same book (e.g., SC -> DC -> SC)
	// should be handled by direct matching, not ring matching
	if len(ring.Orders) == 2 {
		order1 := ring.Orders[0]
		order2 := ring.Orders[1]

		// Check if both orders involve the same two tokens (same book)
		tokens1 := map[string]bool{order1.SymbolIn: true, order1.SymbolOut: true}
		tokens2 := map[string]bool{order2.SymbolIn: true, order2.SymbolOut: true}

		// If both orders only involve the same 2 tokens, it's a same-book ring
		if len(tokens1) == 2 && len(tokens2) == 2 {
			sameTokens := true
			for token := range tokens1 {
				if !tokens2[token] {
					sameTokens = false
					break
				}
			}
			if sameTokens {
				// log.Printf(" Ring rejected: 2-order ring within same book (%s/%s) should use direct matching",
				// 	order1.SymbolIn, order1.SymbolOut)
				return false
			}
		}
	}

	// Ring matching requires at least 3 unique tokens (cross-market arbitrage)
	uniqueTokens := make(map[string]bool)
	for _, order := range ring.Orders {
		uniqueTokens[order.SymbolIn] = true
		uniqueTokens[order.SymbolOut] = true
	}
	if len(uniqueTokens) < 3 {
		// log.Printf(" Ring rejected: only %d unique tokens (need at least 3 for cross-market arbitrage)",
		// 	len(uniqueTokens))
		return false
	}

	// Calculate if the ring has positive value flow
	bottleneck := store.calculateRingBottleneck(ring)
	if bottleneck.Cmp(big.NewInt(0)) <= 0 {
		// log.Printf(" Ring has no tradeable amount")
		return false
	}

	return true
}

func (store *OrderBookStore) calculateRingBottleneck(ring *RingPath) *big.Int {
	if len(ring.Orders) == 0 {
		return big.NewInt(0)
	}

	// log.Printf("  [Bottleneck] Calculating for %d orders in ring", len(ring.Orders))

	// The bottleneck is the maximum amount that can successfully flow through the entire ring
	// We test each order's capacity and find the largest amount that completes the cycle

	var maxValidAmount *big.Int

	for startIdx := 0; startIdx < len(ring.Orders); startIdx++ {
		testAmount := new(big.Int).Set(ring.TradeAmounts[startIdx])

		// log.Printf("  [Bottleneck] Testing with order %d capacity: %s", startIdx, testAmount.String())

		// Simulate flow starting from this order
		currentAmount := new(big.Int).Set(testAmount)
		valid := true

		for i := 0; i < len(ring.Orders); i++ {
			orderIdx := (startIdx + i) % len(ring.Orders)
			order := ring.Orders[orderIdx]

			// Check if this order can handle the current amount
			if currentAmount.Cmp(ring.TradeAmounts[orderIdx]) > 0 {
				// This order can't handle this much
				valid = false
				// log.Printf("  [Bottleneck]    Order %d can't handle %s (max: %s)",
				// orderIdx, currentAmount.String(), ring.TradeAmounts[orderIdx].String())
				break
			}

			// Calculate output from this order
			output := new(big.Int).Mul(currentAmount, order.AmtOut)
			output.Div(output, order.AmtIn)

			// log.Printf("  [Bottleneck]   Order %d: input=%s, output=%s",
			// orderIdx, currentAmount.String(), output.String())

			currentAmount.Set(output)
		}

		// After going through the ring, check if we got back at least what we started with
		if valid && currentAmount.Cmp(testAmount) >= 0 {
			// log.Printf("  [Bottleneck]    Valid flow with amount %s (output: %s)",
			// 	testAmount.String(), currentAmount.String())

			// This is a valid amount - keep the largest one we find
			if maxValidAmount == nil || testAmount.Cmp(maxValidAmount) > 0 {
				maxValidAmount = new(big.Int).Set(testAmount)
				// log.Printf("  [Bottleneck]    New maximum valid amount!")
			}
		} else if valid {
			// log.Printf("  [Bottleneck]    Invalid flow with amount %s (insufficient output: %s)",
			// 	testAmount.String(), currentAmount.String())
		}
	}

	if maxValidAmount == nil {
		// log.Printf("  [Bottleneck] No valid flow found!")
		return big.NewInt(0)
	}

	// log.Printf("  [Bottleneck] Final bottleneck: %s", maxValidAmount.String())
	return maxValidAmount
}

func (store *OrderBookStore) executeRing(ring *RingPath) error {
	bottleneck := store.calculateRingBottleneck(ring)
	if bottleneck.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("invalid ring: no tradeable amount")
	}

	log.Printf("💎 Executing Ring Trade:")
	log.Printf("   Orders: %d", len(ring.Orders))
	log.Printf("   Bottleneck: %s", bottleneck.String())
	log.Printf("   Path: %s", store.ringToString(ring))

	// Lock all books involved
	for _, book := range ring.Books {
		book.Mu.Lock()
		defer book.Mu.Unlock()
	}

	// Find bottleneck index
	bottleneckIdx := -1
	for i, tradeAmount := range ring.TradeAmounts {
		if tradeAmount.Cmp(bottleneck) == 0 {
			bottleneckIdx = i
			log.Printf("   Bottleneck order: #%d (%s)", i+1, getOrderKey(ring.Orders[i])[:20])
			break
		}
	}
	if bottleneckIdx == -1 {
		return fmt.Errorf("could not find bottleneck order")
	}

	// Compute fill amounts based on bottleneck flow
	fillAmounts := make([]*big.Int, len(ring.Orders))
	currentAmount := new(big.Int).Set(bottleneck)
	for i := 0; i < len(ring.Orders); i++ {
		orderIdx := (bottleneckIdx + i) % len(ring.Orders)
		order := ring.Orders[orderIdx]

		fillAmounts[orderIdx] = new(big.Int).Set(currentAmount)

		output := new(big.Int).Mul(currentAmount, order.AmtOut)
		output.Div(output, order.AmtIn)
		currentAmount.Set(output)
	}

	// Log fill percentages before execution
	for i, order := range ring.Orders {
		fillFloat := new(big.Float).SetInt(fillAmounts[i])
		totalFloat := new(big.Float).SetInt(order.AmtIn)
		percentFloat := new(big.Float).Quo(fillFloat, totalFloat)
		percentFloat.Mul(percentFloat, big.NewFloat(100))
		percent, _ := percentFloat.Float64()

		log.Printf("   Before: Order %d: %s/%s | Fill: %s of %s (%.2f%%) | Current: %s/%s",
			i+1, order.CreatedBy.Hex()[:10], order.Nonce.String(),
			fillAmounts[i].String(), order.AmtIn.String(), percent,
			order.FilledAmtIn.String(), order.AmtIn.String())
	}

	// Mark all orders as PENDING before sending transaction
	for _, order := range ring.Orders {
		order.Status = 1
	}

	log.Printf("===SENDING RING TRANSACTION TO BLOCKCHAIN===")
	log.Printf(" Submitting Ring Trade to blockchain")
	if store.Exchange == nil {
		// Reset status on error
		for _, order := range ring.Orders {
			order.Status = 0
		}
		return fmt.Errorf("exchange contract not initialized")
	}

	tx, err := store.Exchange.ExecuteRingTrade(ring.Orders, fillAmounts)
	if err != nil {
		log.Printf("ERROR EXECUTING RING TRADE: %+v", err)
		// Reset status on immediate error
		for _, order := range ring.Orders {
			order.Status = 0
		}
		return fmt.Errorf("on-chain ring trade failed: %w", err)
	}

	txHash := tx.Hash().Hex()
	log.Printf(" Submitted Ring Trade TX: %s", txHash)

	// Notify all users that their orders are pending
	for _, order := range ring.Orders {
		api.NotifyUpdate("TransactionChange", order.CreatedBy, order.ToStringMap())
	}

	// Create copies of all data needed in goroutine to avoid race conditions
	finalOrders := make([]*order.Order, len(ring.Orders))
	finalBooks := make([]*MarketOrderBook, len(ring.Books))
	finalPriceLevels := make([]*PriceLevel, len(ring.PriceLevels))
	finalElements := make([]*list.Element, len(ring.Elements))
	finalFillAmounts := make([]*big.Int, len(fillAmounts))

	for i := range ring.Orders {
		finalOrders[i] = ring.Orders[i]
		finalBooks[i] = ring.Books[i]
		finalPriceLevels[i] = ring.PriceLevels[i]
		finalElements[i] = ring.Elements[i]
		finalFillAmounts[i] = new(big.Int).Set(fillAmounts[i])
	}

	// Launch async goroutine to wait for confirmation
	go func() {
		// Wait for transaction receipt
		minedStatus := make(chan int)
		go store.Exchange.Client.CheckTxReceipt(context.Background(), tx, minedStatus)

		result := <-minedStatus

		// Re-acquire locks for all books involved
		for _, book := range finalBooks {
			book.Mu.Lock()
			defer book.Mu.Unlock()
		}

		if result == 1 { // Success
			log.Printf(" Ring Transaction %s confirmed", txHash)

			// Process each order in the ring
			for i, order := range finalOrders {
				// Add transaction hash to order
				order.TransactionHashes = append(order.TransactionHashes, txHash)

				// Update filled amounts (CUMULATIVE)
				order.FilledAmtIn.Add(order.FilledAmtIn, finalFillAmounts[i])

				remaining := new(big.Int).Sub(order.AmtIn, order.FilledAmtIn)

				log.Printf("   After: Order %d: %s/%s | Filled: %s/%s (%.1f%%)",
					i+1, order.CreatedBy.Hex()[:10], order.Nonce.String(),
					order.FilledAmtIn.String(), order.AmtIn.String(),
					percent(order.FilledAmtIn, order.AmtIn))

				// Update price level total quantity
				finalPriceLevels[i].TotalQuantity.Sub(
					finalPriceLevels[i].TotalQuantity,
					finalFillAmounts[i],
				)

				if remaining.Cmp(big.NewInt(0)) == 0 {
					// Fully filled - status 3, add to history, then set to 2
					order.Status = 3
					store.AddToPastHistory(order)

					finalPriceLevels[i].Orders.Remove(finalElements[i])
					order.Status = 2 // Fully filled
					log.Printf("    Order %s/%s fully filled (100%%) - Added to history",
						order.CreatedBy.Hex()[:10], order.Nonce.String())

					// Handle conditional order
					if order.ConditionalOrder != nil {
						log.Printf("**Conditional Order Detected**: Storing conditional order for %s/%s",
							order.CreatedBy.Hex()[:10], order.Nonce.String())
						parentOrderID := fmt.Sprintf("%s-%s", order.CreatedBy.Hex(), order.Nonce.String())
						conditionalOrderCopy := order.ConditionalOrder
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
					order.Status = 5
					store.AddToPastHistory(order)

					order.Status = 0 // Back to active
					log.Printf("    Order %s/%s partially filled - Added to history",
						order.CreatedBy.Hex()[:10], order.Nonce.String())
				}

				// Notify user of order update
				api.NotifyUpdate("TransactionChange", order.CreatedBy, order.ToStringMap())
			}

			// 🧹 Remove empty price levels and notify orderbook updates
			processedBooks := make(map[*MarketOrderBook]bool)
			for i, book := range finalBooks {
				if finalPriceLevels[i].Orders.Len() == 0 {
					// Find and remove the empty price level
					iter := book.Asks.Iterator()
					for iter.Next() {
						if iter.Value().(*PriceLevel) == finalPriceLevels[i] {
							priceKey := iter.Key().(*big.Int)
							book.Asks.Remove(priceKey)
							log.Printf("   Removed empty ask price level from %s/%s", book.SymbolOut, book.SymbolIn)
							break
						}
					}
					iter = book.Bids.Iterator()
					for iter.Next() {
						if iter.Value().(*PriceLevel) == finalPriceLevels[i] {
							priceKey := iter.Key().(*big.Int)
							book.Bids.Remove(priceKey)
							log.Printf("   Removed empty bid price level from %s/%s", book.SymbolOut, book.SymbolIn)
							break
						}
					}
				}

				// Send orderbook update notification (only once per book)
				if !processedBooks[book] {
					book.NotifyUpdate("RingMatch", book.Snapshot())
					processedBooks[book] = true
				}
			}

			log.Printf(" Ring execution complete (TX: %s)!", txHash)
		} else {
			log.Printf(" Ring Transaction %s failed or reverted", txHash)
			// Reset all orders to active status
			for _, order := range finalOrders {
				order.Status = 0
				api.NotifyUpdate("TransactionChange", order.CreatedBy, order.ToStringMap())
			}
		}
	}()

	return nil
}

func getOrderKey(o *order.Order) string {
	return fmt.Sprintf("%s-%s", o.CreatedBy.Hex(), o.Nonce.String())
}

func (store *OrderBookStore) ringToString(ring *RingPath) string {
	if len(ring.Orders) == 0 {
		return ""
	}

	result := ring.Orders[0].SymbolIn
	for _, order := range ring.Orders {
		result += fmt.Sprintf(" -> %s", order.SymbolOut)
	}
	return result
}

func (store *OrderBookStore) InitializeBook(tokenA, tokenB string) {
	store.mu.Lock()
	defer store.mu.Unlock()

	base, quote := GetPairKey(tokenA, tokenB)
	pairID := base + "/" + quote

	if _, exists := store.Books[pairID]; !exists {
		store.Books[pairID] = NewMarketOrderBook(base, quote)
		log.Printf("**Book Initialized**: %s (BASE: %s, QUOTE: %s)", pairID, base, quote)
	} else {
		log.Printf("**Book Exists**: %s already initialized", pairID)
	}
}

func (store *OrderBookStore) AddOrder(orderIn *order.Order) error {
	base, quote := GetPairKey(orderIn.SymbolIn, orderIn.SymbolOut)
	pairID := base + "/" + quote

	store.mu.RLock()
	book, exists := store.Books[pairID]
	store.mu.RUnlock()

	if !exists {
		log.Printf("**Order Rejected**: Book for %s not initialized", pairID)
		return fmt.Errorf("order book for %s not initialized", pairID)
	}

	orderId := orderIn.CreatedBy.String() + "/" + orderIn.Nonce.String()

	book.Mu.Lock()
	defer book.Mu.Unlock()

	// Initialize FilledAmtIn if not set (new orders start with 0 filled)
	if orderIn.FilledAmtIn == nil {
		orderIn.FilledAmtIn = big.NewInt(0)
	}

	var isBid bool
	var priceKey *big.Int
	var side string

	if orderIn.SymbolIn == base && orderIn.SymbolOut == quote {
		// SELL/ASK: Giving base, wanting quote
		// Price = how much quote they want per base = AmtOut / AmtIn
		isBid = false
		side = "ASK"
		priceBigFloat := new(big.Float).Quo(
			new(big.Float).SetInt(orderIn.AmtOut),
			new(big.Float).SetInt(orderIn.AmtIn),
		)
		priceBigFloat.Mul(priceBigFloat, new(big.Float).SetInt(PriceFactor))
		priceKey = new(big.Int)
		priceBigFloat.Int(priceKey)

	} else if orderIn.SymbolIn == quote && orderIn.SymbolOut == base {
		// BUY/BID: Giving quote, wanting base
		// Price = how much quote they're paying per base = AmtIn / AmtOut
		isBid = true
		side = "BID"
		priceBigFloat := new(big.Float).Quo(
			new(big.Float).SetInt(orderIn.AmtIn),
			new(big.Float).SetInt(orderIn.AmtOut),
		)
		priceBigFloat.Mul(priceBigFloat, new(big.Float).SetInt(PriceFactor))
		priceKey = new(big.Int)
		priceBigFloat.Int(priceKey)

	} else {
		return fmt.Errorf("invalid order: tokens don't match book %s/%s", base, quote)
	}

	// Store the price key on the order
	orderIn.LimitPrice = priceKey

	priceFloat := new(big.Float).Quo(new(big.Float).SetInt(priceKey), new(big.Float).SetInt(PriceFactor))
	priceStr, _ := priceFloat.Float64()

	// Calculate remaining amount for display
	remainingIn := new(big.Int).Sub(orderIn.AmtIn, orderIn.FilledAmtIn)
	fillPercent := 0.0
	if orderIn.AmtIn.Cmp(big.NewInt(0)) > 0 {
		fillPercent = float64(orderIn.FilledAmtIn.Int64()) / float64(orderIn.AmtIn.Int64()) * 100
	}

	log.Printf("**Order Added**: %s | %s | Price: %.6f | Original: %s %s -> %s %s | Filled: %.2f%% | Remaining: %s",
		orderId[:20], side, priceStr,
		orderIn.AmtIn.String(), orderIn.SymbolIn,
		orderIn.AmtOut.String(), orderIn.SymbolOut,
		fillPercent, remainingIn.String())

	// Select the correct tree
	var tree *rbtree.Tree
	if isBid {
		tree = book.Bids
	} else {
		tree = book.Asks
	}

	// Add order to the appropriate price level
	// Note: TotalQuantity should track remaining amounts, not original
	val, found := tree.Get(priceKey)
	if !found {
		pl := &PriceLevel{
			Orders:        list.New(),
			TotalQuantity: new(big.Int).Set(remainingIn), // Use remaining, not original
		}
		pl.Orders.PushBack(orderIn)
		tree.Put(priceKey, pl)
	} else {
		pl := val.(*PriceLevel)
		pl.Orders.PushBack(orderIn)
		pl.TotalQuantity.Add(pl.TotalQuantity, remainingIn) // Add remaining, not original
	}
	book.NotifyUpdate("Add", book.Snapshot())
	api.NotifyUpdate("OrderAdd", orderIn.CreatedBy, orderIn.ToStringMap())
	return nil
}

// StoreConditionalOrder stores a conditional order for later execution when price conditions are met
func (store *OrderBookStore) StoreConditionalOrder(conditionalOrder *order.Order, parentOrderID string) error {
	if conditionalOrder == nil {
		return fmt.Errorf("conditional order is nil")
	}

	// Prevent recursive conditional orders
	if conditionalOrder.ConditionalOrder != nil {
		log.Printf("**Warning**: Conditional order has its own conditional order - removing to prevent chain")
		conditionalOrder.ConditionalOrder = nil
	}

	// Validate that TriggerPrice is set
	if conditionalOrder.TriggerPrice == nil || conditionalOrder.TriggerPrice.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("conditional order has invalid trigger price: %v", conditionalOrder.TriggerPrice)
	}

	// For now, we only support STOP_LIMIT type
	// In the future, this could be passed from the frontend via conditionalOrder.ConditionalType (if ever)
	conditionalType := ConditionalTypeStopLimit

	// Determine trigger direction based on the order type
	// For a stop-loss sell: if I'm selling token A to get token B, trigger when A/B price <= triggerPrice
	// For a stop-loss buy: if I'm buying token A with token B, trigger when A/B price >= triggerPrice

	base, quote := GetPairKey(conditionalOrder.SymbolIn, conditionalOrder.SymbolOut)
	var isTriggerAbove bool

	if conditionalOrder.SymbolIn == base && conditionalOrder.SymbolOut == quote {
		// SELL/ASK: Giving base, wanting quote
		// This is a stop-loss sell - trigger when price drops to or below this level
		isTriggerAbove = false // Trigger when price <= triggerPrice

	} else if conditionalOrder.SymbolIn == quote && conditionalOrder.SymbolOut == base {
		// BUY/BID: Giving quote, wanting base
		// This is a stop-loss buy - trigger when price rises to or above this level
		isTriggerAbove = true // Trigger when price >= triggerPrice

	} else {
		return fmt.Errorf("invalid conditional order: tokens don't match any book")
	}

	// Store in the conditional order store with trigger conditions
	return store.ConditionalOrderStore.AddConditionalOrder(
		conditionalOrder,
		parentOrderID,
		conditionalType,
		base,
		quote,
		isTriggerAbove,
	)
}

// RemoveOrder removes an order from the order book using minimal identifiers
func (store *OrderBookStore) RemoveOrder(createdBy common.Address, nonce *big.Int, limitPrice *big.Int, tokenA, tokenB string) {
	base, quote := GetPairKey(tokenA, tokenB)
	pairID := base + "/" + quote

	store.mu.RLock()
	book, exists := store.Books[pairID]
	store.mu.RUnlock()

	if !exists {
		log.Printf("**Order Rejected**: Book for %s not initialized", pairID)
		return
	}

	orderId := createdBy.String() + "/" + nonce.String()
	log.Printf("**Order Removal**: Removing %s from %s at price %s", orderId, pairID, limitPrice.String())

	book.Mu.Lock()
	defer book.Mu.Unlock()

	if limitPrice == nil {
		log.Printf("**Error**: Order %s has no price key", orderId)
		return
	}

	// Try to find the order in both bids and asks
	// We'll search both since we dk which side without full order details
	var foundOrder *order.Order
	var foundElem *list.Element
	var foundLevel *PriceLevel
	var foundTree *rbtree.Tree

	// Search in bids
	val, found := book.Bids.Get(limitPrice)
	if found {
		pl := val.(*PriceLevel)
		for e := pl.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.CreatedBy.Cmp(createdBy) == 0 && o.Nonce.Cmp(nonce) == 0 {
				foundOrder = o
				foundElem = e
				foundLevel = pl
				foundTree = book.Bids
				break
			}
		}
	}

	// If not found in bids, search in asks
	if foundOrder == nil {
		val, found := book.Asks.Get(limitPrice)
		if found {
			pl := val.(*PriceLevel)
			for e := pl.Orders.Front(); e != nil; e = e.Next() {
				o := e.Value.(*order.Order)
				if o.CreatedBy.Cmp(createdBy) == 0 && o.Nonce.Cmp(nonce) == 0 {
					foundOrder = o
					foundElem = e
					foundLevel = pl
					foundTree = book.Asks
					break
				}
			}
		}
	}

	if foundOrder == nil {
		log.Printf("**Warning**: Order %s not found at price level %s", orderId, limitPrice.String())
		return
	}

	// Set status to 4 (cancelled) and add to history
	foundOrder.Status = 4
	store.AddToPastHistory(foundOrder)

	// Calculate remaining amount to remove from TotalQuantity
	if foundOrder.FilledAmtIn == nil {
		foundOrder.FilledAmtIn = big.NewInt(0)
	}
	quantityToRemove := new(big.Int).Sub(foundOrder.AmtIn, foundOrder.FilledAmtIn)

	// Remove the order from the list
	foundLevel.Orders.Remove(foundElem)

	log.Printf("**Order Removed**: %s from %s (was %s/%s filled) - Added to history with status 4 (cancelled)",
		orderId, pairID, foundOrder.FilledAmtIn.String(), foundOrder.AmtIn.String())

	// Update total quantity (remove the remaining amount, not original)
	foundLevel.TotalQuantity.Sub(foundLevel.TotalQuantity, quantityToRemove)

	// Remove empty price levels
	if foundLevel.Orders.Len() == 0 {
		foundTree.Remove(limitPrice)
		log.Printf("   Removed empty price level %s from %s", limitPrice.String(), pairID)
	}

	book.NotifyUpdate("Remove", book.Snapshot())
	api.NotifyUpdate("OrderRemove", createdBy, map[string]any{"nonce": nonce})
}

// GetOrdersByCreator returns all orders for a given creator across all books
func (store *OrderBookStore) GetOrdersByCreator(creator common.Address) []*order.Order {
	store.mu.RLock()
	books := make([]*MarketOrderBook, 0, len(store.Books))
	for _, book := range store.Books {
		books = append(books, book)
	}
	store.mu.RUnlock()

	var orders []*order.Order

	for _, book := range books {

		bookOrders := []*order.Order{}
		book.Mu.RLock()

		collect := func(tree *rbtree.Tree) {
			it := tree.Iterator()
			for it.Next() {
				level := it.Value().(*PriceLevel)
				for e := level.Orders.Front(); e != nil; e = e.Next() {
					o := e.Value.(*order.Order)
					if o.CreatedBy == creator {
						orders = append(orders, o)
						bookOrders = append(bookOrders, o)
					}
				}
			}
		}

		collect(book.Bids)
		collect(book.Asks)

		book.Mu.RUnlock()
		log.Printf("**Query**: Found %d total orders for creator in %s", len(bookOrders), book.SymbolIn+"/"+book.SymbolOut)
	}

	log.Printf("**Query Complete**: Found %d total orders for creator", len(orders))
	return orders
}

// GetOrdersByCreatorInBook returns all orders for a given creator in a specific book
func (store *OrderBookStore) GetOrdersByCreatorInBook(creatorAddress common.Address, tokenA, tokenB string) ([]*order.Order, error) {
	base, quote := GetPairKey(tokenA, tokenB)
	pairID := base + "/" + quote

	store.mu.RLock()
	book, exists := store.Books[pairID]
	store.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("order book for %s not found", pairID)
	}

	book.Mu.RLock()
	defer book.Mu.RUnlock()

	var orders []*order.Order

	// Search bids
	it := book.Bids.Iterator()
	for it.Next() {
		priceLevel := it.Value().(*PriceLevel)
		for e := priceLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.CreatedBy.Cmp(creatorAddress) == 0 {
				orders = append(orders, o)
			}
		}
	}

	// Search asks
	it = book.Asks.Iterator()
	for it.Next() {
		priceLevel := it.Value().(*PriceLevel)
		for e := priceLevel.Orders.Front(); e != nil; e = e.Next() {
			o := e.Value.(*order.Order)
			if o.CreatedBy.Cmp(creatorAddress) == 0 {
				orders = append(orders, o)
			}
		}
	}

	log.Printf("**Query**: Found %d orders for creator in %s", len(orders), pairID)
	return orders, nil
}

// GetMarketPrice returns the current market price for a token pair
func (store *OrderBookStore) GetMarketPrice(tokenA, tokenB string) (*MarketPrice, error) {
	base, quote := GetPairKey(tokenA, tokenB)
	pairID := base + "/" + quote

	store.mu.RLock()
	book, exists := store.Books[pairID]
	store.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("order book for %s not found", pairID)
	}

	book.Mu.RLock()
	defer book.Mu.RUnlock()

	price := &MarketPrice{}

	// Get best bid (highest price)
	if bestBidNode := book.Bids.Left(); bestBidNode != nil {
		price.BestBid = bestBidNode.Key.(*big.Int)
	}

	// Get best ask (lowest price)
	if bestAskNode := book.Asks.Left(); bestAskNode != nil {
		price.BestAsk = bestAskNode.Key.(*big.Int)
	}

	// Calculate mid price and spread if both sides exist
	if price.BestBid != nil && price.BestAsk != nil {
		price.MidPrice = new(big.Int).Add(price.BestBid, price.BestAsk)
		price.MidPrice.Div(price.MidPrice, big.NewInt(2))

		price.Spread = new(big.Int).Sub(price.BestAsk, price.BestBid)
	}

	return price, nil
}

// GetPriceFloat returns prices as float64 for display purposes
func (store *OrderBookStore) GetPriceFloat(tokenA, tokenB string) (map[string]float64, error) {
	price, err := store.GetMarketPrice(tokenA, tokenB)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)

	if price.BestBid != nil {
		bidFloat := new(big.Float).Quo(
			new(big.Float).SetInt(price.BestBid),
			new(big.Float).SetInt(PriceFactor),
		)
		result["best_bid"], _ = bidFloat.Float64()
	}

	if price.BestAsk != nil {
		askFloat := new(big.Float).Quo(
			new(big.Float).SetInt(price.BestAsk),
			new(big.Float).SetInt(PriceFactor),
		)
		result["best_ask"], _ = askFloat.Float64()
	}

	if price.MidPrice != nil {
		midFloat := new(big.Float).Quo(
			new(big.Float).SetInt(price.MidPrice),
			new(big.Float).SetInt(PriceFactor),
		)
		result["mid_price"], _ = midFloat.Float64()
	}

	if price.Spread != nil {
		spreadFloat := new(big.Float).Quo(
			new(big.Float).SetInt(price.Spread),
			new(big.Float).SetInt(PriceFactor),
		)
		result["spread"], _ = spreadFloat.Float64()
	}

	return result, nil
}

// GetOrderBookSnapshot returns top depth levels of bids and asks with prices
func (store *OrderBookStore) GetOrderBookSnapshot(tokenA, tokenB string, depth int) (map[string]any, error) {
	base, quote := GetPairKey(tokenA, tokenB)
	pairID := base + "/" + quote

	store.mu.RLock()
	book, exists := store.Books[pairID]
	store.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("order book for %s not found", pairID)
	}

	book.Mu.RLock()
	defer book.Mu.RUnlock()

	snapshot := make(map[string]any)
	snapshot["pair"] = pairID
	snapshot["base"] = base
	snapshot["quote"] = quote

	bids := []map[string]string{}
	bidIterator := book.Bids.Iterator()
	count := 0
	for bidIterator.Next() && count < depth {
		priceKey := bidIterator.Key().(*big.Int)
		level := bidIterator.Value().(*PriceLevel)

		priceFloat := new(big.Float).Quo(
			new(big.Float).SetInt(priceKey),
			new(big.Float).SetInt(PriceFactor),
		)
		priceStr, _ := priceFloat.Float64()

		bids = append(bids, map[string]string{
			"price":    fmt.Sprintf("%.6f", priceStr),
			"quantity": level.TotalQuantity.String(),
			"orders":   fmt.Sprintf("%d", level.Orders.Len()),
		})
		count++
	}
	snapshot["bids"] = bids

	asks := []map[string]string{}
	askIterator := book.Asks.Iterator()
	count = 0
	for askIterator.Next() && count < depth {
		priceKey := askIterator.Key().(*big.Int)
		level := askIterator.Value().(*PriceLevel)

		priceFloat := new(big.Float).Quo(
			new(big.Float).SetInt(priceKey),
			new(big.Float).SetInt(PriceFactor),
		)
		priceStr, _ := priceFloat.Float64()

		asks = append(asks, map[string]string{
			"price":    fmt.Sprintf("%.6f", priceStr),
			"quantity": level.TotalQuantity.String(),
			"orders":   fmt.Sprintf("%d", level.Orders.Len()),
		})
		count++
	}
	snapshot["asks"] = asks

	return snapshot, nil
}
