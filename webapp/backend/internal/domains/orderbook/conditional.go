package orderbook

import (
	"context"
	"dexbe/internal/domains/order"
	"fmt"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

// ConditionalOrderStore manages conditional orders and monitors their trigger conditions
type ConditionalOrderStore struct {
	// Map: creator address + nonce -> conditional order
	conditionalOrders map[string]*ConditionalOrderEntry
	mu                sync.RWMutex
	orderBookStore    *OrderBookStore
}

// ConditionalType represents the type of conditional order
type ConditionalType string

const (
	ConditionalTypeStopLimit ConditionalType = "STOP_LIMIT"
)

// ConditionalOrderEntry wraps a conditional order with its trigger condition
type ConditionalOrderEntry struct {
	Order            *order.Order
	ParentOrderID    string // Format: "address-nonce" of the parent order
	ConditionalType  ConditionalType
	TriggerPrice     *big.Int
	TriggerSymbolIn  string // The token pair to watch for price
	TriggerSymbolOut string
	IsTriggerAbove   bool // true = trigger when price >= triggerPrice, false = trigger when price <= triggerPrice
	CreatedAt        time.Time
}

func NewConditionalOrderStore(orderBookStore *OrderBookStore) *ConditionalOrderStore {
	return &ConditionalOrderStore{
		conditionalOrders: make(map[string]*ConditionalOrderEntry),
		orderBookStore:    orderBookStore,
	}
}

// AddConditionalOrder stores a conditional order when its parent order is filled
func (store *ConditionalOrderStore) AddConditionalOrder(
	conditionalOrder *order.Order,
	parentOrderID string,
	conditionalType ConditionalType,
	triggerSymbolIn string,
	triggerSymbolOut string,
	isTriggerAbove bool,
) error {
	if conditionalOrder == nil {
		return fmt.Errorf("conditional order is nil")
	}

	// Validate the conditional order
	if conditionalOrder.SymbolIn == "" || conditionalOrder.SymbolOut == "" {
		return fmt.Errorf("conditional order has invalid symbols (In: %s, Out: %s)",
			conditionalOrder.SymbolIn, conditionalOrder.SymbolOut)
	}

	if conditionalOrder.AmtIn == nil || conditionalOrder.AmtIn.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("conditional order has invalid AmtIn: %v", conditionalOrder.AmtIn)
	}

	if conditionalOrder.AmtOut == nil || conditionalOrder.AmtOut.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("conditional order has invalid AmtOut: %v", conditionalOrder.AmtOut)
	}

	if conditionalOrder.Nonce == nil {
		return fmt.Errorf("conditional order has nil Nonce")
	}

	if conditionalOrder.TriggerPrice == nil || conditionalOrder.TriggerPrice.Cmp(big.NewInt(0)) <= 0 {
		return fmt.Errorf("invalid trigger price: %v", conditionalOrder.TriggerPrice)
	}

	// Validate conditional type
	if conditionalType != ConditionalTypeStopLimit {
		return fmt.Errorf("unsupported conditional type: %s (only STOP_LIMIT is currently supported)", conditionalType)
	}

	orderKey := getConditionalOrderKey(conditionalOrder.CreatedBy, conditionalOrder.Nonce)

	store.mu.Lock()
	defer store.mu.Unlock()

	// Check if already exists
	if _, exists := store.conditionalOrders[orderKey]; exists {
		return fmt.Errorf("conditional order already exists: %s", orderKey)
	}

	entry := &ConditionalOrderEntry{
		Order:            conditionalOrder,
		ParentOrderID:    parentOrderID,
		ConditionalType:  conditionalType,
		TriggerPrice:     new(big.Int).Set(conditionalOrder.TriggerPrice),
		TriggerSymbolIn:  triggerSymbolIn,
		TriggerSymbolOut: triggerSymbolOut,
		IsTriggerAbove:   isTriggerAbove,
		CreatedAt:        time.Now(),
	}

	store.conditionalOrders[orderKey] = entry

	// Log the trigger price in readable format
	triggerPriceFloat := new(big.Float).Quo(
		new(big.Float).SetInt(conditionalOrder.TriggerPrice),
		new(big.Float).SetInt(PriceFactor),
	)
	priceFloat, _ := triggerPriceFloat.Float64()
	triggerType := "below or equal"
	if isTriggerAbove {
		triggerType = "above or equal"
	}

	log.Printf("**Conditional Order Stored**: %s/%s | Type: %s | Parent: %s | Trigger: %s/%s %s %.6f | Order: %s %s -> %s %s",
		conditionalOrder.CreatedBy.Hex()[:10],
		conditionalOrder.Nonce.String(),
		conditionalType,
		parentOrderID,
		triggerSymbolIn,
		triggerSymbolOut,
		triggerType,
		priceFloat,
		conditionalOrder.AmtIn.String(),
		conditionalOrder.SymbolIn,
		conditionalOrder.AmtOut.String(),
		conditionalOrder.SymbolOut,
	)

	return nil
}

// RemoveConditionalOrder removes a conditional order (e.g., if manually cancelled)
func (store *ConditionalOrderStore) RemoveConditionalOrder(creator common.Address, nonce *big.Int) error {
	orderKey := getConditionalOrderKey(creator, nonce)

	store.mu.Lock()
	defer store.mu.Unlock()

	entry, exists := store.conditionalOrders[orderKey]
	if !exists {
		return fmt.Errorf("conditional order not found: %s", orderKey)
	}

	delete(store.conditionalOrders, orderKey)

	log.Printf("**Conditional Order Removed**: %s/%s",
		creator.Hex()[:10], nonce.String())

	_ = entry // Avoid unused variable warning
	return nil
}

// GetConditionalOrder retrieves a conditional order
func (store *ConditionalOrderStore) GetConditionalOrder(creator common.Address, nonce *big.Int) (*ConditionalOrderEntry, error) {
	orderKey := getConditionalOrderKey(creator, nonce)

	store.mu.RLock()
	defer store.mu.RUnlock()

	entry, exists := store.conditionalOrders[orderKey]
	if !exists {
		return nil, fmt.Errorf("conditional order not found: %s", orderKey)
	}

	return entry, nil
}

// GetAllConditionalOrders returns all stored conditional orders
func (store *ConditionalOrderStore) GetAllConditionalOrders() []*ConditionalOrderEntry {
	store.mu.RLock()
	defer store.mu.RUnlock()

	entries := make([]*ConditionalOrderEntry, 0, len(store.conditionalOrders))
	for _, entry := range store.conditionalOrders {
		entries = append(entries, entry)
	}

	return entries
}

// GetConditionalOrdersByCreator returns all conditional orders for a given creator
func (store *ConditionalOrderStore) GetConditionalOrdersByCreator(creator common.Address) []*ConditionalOrderEntry {
	store.mu.RLock()
	defer store.mu.RUnlock()

	entries := []*ConditionalOrderEntry{}
	for _, entry := range store.conditionalOrders {
		if entry.Order.CreatedBy == creator {
			entries = append(entries, entry)
		}
	}

	return entries
}

// StartOracle starts monitoring conditional orders and triggers them when conditions are met
func (store *ConditionalOrderStore) StartOracle(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Println("**Conditional Order Oracle Started**: Monitoring trigger conditions")
		for {
			select {
			case <-ctx.Done():
				log.Println("**Conditional Order Oracle Stopped**")
				return
			case <-ticker.C:
				store.checkAndTriggerConditionalOrders()
			}
		}
	}()
}

// checkAndTriggerConditionalOrders checks all conditional orders and triggers those whose conditions are met
func (store *ConditionalOrderStore) checkAndTriggerConditionalOrders() {
	store.mu.RLock()
	entriesToCheck := make([]*ConditionalOrderEntry, 0, len(store.conditionalOrders))
	for _, entry := range store.conditionalOrders {
		entriesToCheck = append(entriesToCheck, entry)
	}
	store.mu.RUnlock()

	if len(entriesToCheck) == 0 {
		return
	}

	// log.Printf("**Conditional Oracle**: Checking %d conditional orders", len(entriesToCheck))

	for _, entry := range entriesToCheck {
		// Handle different conditional types
		shouldTrigger := false
		var triggerReason string

		switch entry.ConditionalType {
		case ConditionalTypeStopLimit:
			shouldTrigger, triggerReason = store.checkStopLimitCondition(entry)
		default:
			log.Printf("**Warning**: Unknown conditional type: %s", entry.ConditionalType)
			continue
		}

		if shouldTrigger {
			// Acquire write lock and double-check entry still exists
			store.mu.Lock()
			orderKey := getConditionalOrderKey(entry.Order.CreatedBy, entry.Order.Nonce)

			if _, exists := store.conditionalOrders[orderKey]; !exists {
				// Already processed by another goroutine
				store.mu.Unlock()
				continue
			}

			// Remove from store BEFORE adding to order book to prevent re-triggering
			delete(store.conditionalOrders, orderKey)
			store.mu.Unlock()

			log.Printf("**Conditional Order TRIGGERED**: %s/%s | Type: %s | Reason: %s",
				entry.Order.CreatedBy.Hex()[:10],
				entry.Order.Nonce.String(),
				entry.ConditionalType,
				triggerReason,
			)

			// Add the order to the order book
			err := store.orderBookStore.AddOrder(entry.Order)
			if err != nil {
				log.Printf("**Error Adding Triggered Conditional Order**: %v", err)

				// Re-add to conditional store on error so it can be retried
				store.mu.Lock()
				store.conditionalOrders[orderKey] = entry
				store.mu.Unlock()

				log.Printf("**Conditional Order Re-queued**: %s/%s (will retry on next check)",
					entry.Order.CreatedBy.Hex()[:10],
					entry.Order.Nonce.String(),
				)
				continue
			}

			log.Printf("**Conditional Order Added to Book**: %s/%s | %s %s -> %s %s",
				entry.Order.CreatedBy.Hex()[:10],
				entry.Order.Nonce.String(),
				entry.Order.AmtIn.String(),
				entry.Order.SymbolIn,
				entry.Order.AmtOut.String(),
				entry.Order.SymbolOut,
			)
		}
	}
}

// checkStopLossCondition checks if a stop-loss condition is met
func (store *ConditionalOrderStore) checkStopLimitCondition(entry *ConditionalOrderEntry) (bool, string) {
	// Get current market price for the trigger pair
	marketPrice, err := store.orderBookStore.GetMarketPrice(entry.TriggerSymbolIn, entry.TriggerSymbolOut)
	if err != nil {
		// log.Printf("**Warning**: Could not get market price for %s/%s: %v",
		// 	entry.TriggerSymbolIn, entry.TriggerSymbolOut, err)
		return false, ""
	}

	// Use mid price as the current price
	if marketPrice.MidPrice == nil {
		// log.Printf("**Warning**: No mid price available for %s/%s",
		// 	entry.TriggerSymbolIn, entry.TriggerSymbolOut)
		return false, ""
	}

	currentPrice := marketPrice.MidPrice

	// Check if trigger condition is met
	shouldTrigger := false
	if entry.IsTriggerAbove {
		// Trigger when price >= triggerPrice
		shouldTrigger = currentPrice.Cmp(entry.TriggerPrice) >= 0
	} else {
		// Trigger when price <= triggerPrice
		shouldTrigger = currentPrice.Cmp(entry.TriggerPrice) <= 0
	}

	if shouldTrigger {
		// Format trigger reason
		currentPriceFloat := new(big.Float).Quo(
			new(big.Float).SetInt(currentPrice),
			new(big.Float).SetInt(PriceFactor),
		)
		triggerPriceFloat := new(big.Float).Quo(
			new(big.Float).SetInt(entry.TriggerPrice),
			new(big.Float).SetInt(PriceFactor),
		)
		cpf, _ := currentPriceFloat.Float64()
		tpf, _ := triggerPriceFloat.Float64()

		reason := fmt.Sprintf("%s/%s: %.6f %s %.6f",
			entry.TriggerSymbolIn,
			entry.TriggerSymbolOut,
			cpf,
			map[bool]string{true: ">=", false: "<="}[entry.IsTriggerAbove],
			tpf,
		)

		return true, reason
	}

	return false, ""
}

// Helper function to create a unique key for conditional orders
func getConditionalOrderKey(creator common.Address, nonce *big.Int) string {
	return fmt.Sprintf("%s-%s", creator.Hex(), nonce.String())
}

// GetConditionalOrderCount returns the number of conditional orders currently stored
func (store *ConditionalOrderStore) GetConditionalOrderCount() int {
	store.mu.RLock()
	defer store.mu.RUnlock()
	return len(store.conditionalOrders)
}

// CheckPriceTriggersForBook checks if any conditional orders should be triggered
// based on the new last price for a specific book
func (store *ConditionalOrderStore) CheckPriceTriggersForBook(symbolIn, symbolOut string, lastPrice *big.Int) {
	if lastPrice == nil || lastPrice.Cmp(big.NewInt(0)) <= 0 {
		return
	}

	// The book's price is always base/quote where symbolIn is base, symbolOut is quote
	// lastPrice = how much quote per base
	bookBase := symbolIn
	bookQuote := symbolOut

	store.mu.RLock()
	entriesToCheck := make([]*ConditionalOrderEntry, 0)
	for _, entry := range store.conditionalOrders {
		// Normalize both the trigger pair and book pair
		entryBase, entryQuote := GetPairKey(entry.TriggerSymbolIn, entry.TriggerSymbolOut)
		normalizedBookBase, normalizedBookQuote := GetPairKey(bookBase, bookQuote)

		// Only check orders watching this pair (in either direction)
		if entryBase == normalizedBookBase && entryQuote == normalizedBookQuote {
			entriesToCheck = append(entriesToCheck, entry)
		}
	}
	store.mu.RUnlock()

	if len(entriesToCheck) == 0 {
		return
	}

	log.Printf(" Checking %d conditional orders for %s/%s at last price %s",
		len(entriesToCheck), bookBase, bookQuote, lastPrice.String())

	for _, entry := range entriesToCheck {
		log.Printf(" Entry: TriggerSymbolIn=%s, TriggerSymbolOut=%s, TriggerPrice=%s, IsTriggerAbove=%v",
			entry.TriggerSymbolIn, entry.TriggerSymbolOut, entry.TriggerPrice.String(), entry.IsTriggerAbove)

		// Calculate the price in the direction the conditional order is watching
		var priceToCheck *big.Int

		// Determine if we need to invert the price
		if entry.TriggerSymbolIn == bookBase && entry.TriggerSymbolOut == bookQuote {
			// Same direction: trigger is watching base/quote, same as book
			priceToCheck = new(big.Int).Set(lastPrice)
			log.Printf(" Same direction: priceToCheck = lastPrice = %s", priceToCheck.String())
		} else if entry.TriggerSymbolIn == bookQuote && entry.TriggerSymbolOut == bookBase {
			// Opposite direction: trigger is watching quote/base, need to invert
			// inverted price = 1 / lastPrice = PriceFactor^2 / lastPrice
			priceToCheck = new(big.Int).Mul(PriceFactor, PriceFactor)
			priceToCheck.Div(priceToCheck, lastPrice)
			log.Printf(" Opposite direction: priceToCheck = inverted = %s", priceToCheck.String())
		} else {
			// This shouldn't happen due to the filter above but jic
			log.Printf(" Unexpected token mismatch for conditional order %s/%s",
				entry.Order.CreatedBy.Hex()[:10], entry.Order.Nonce.String())
			continue
		}

		shouldTrigger := false
		var triggerReason string

		// Check trigger condition based on price
		if entry.IsTriggerAbove {
			// Trigger when price >= triggerPrice
			shouldTrigger = priceToCheck.Cmp(entry.TriggerPrice) >= 0
			log.Printf(" IsTriggerAbove=true: %s >= %s? Result: %v",
				priceToCheck.String(), entry.TriggerPrice.String(), shouldTrigger)
		} else {
			// Trigger when price <= triggerPrice
			shouldTrigger = priceToCheck.Cmp(entry.TriggerPrice) <= 0
			log.Printf(" IsTriggerAbove=false: %s <= %s? Result: %v",
				priceToCheck.String(), entry.TriggerPrice.String(), shouldTrigger)
		}

		if shouldTrigger {
			log.Printf(" Condition met! Attempting to trigger...")

			// Acquire write lock and double-check entry still exists
			store.mu.Lock()
			orderKey := getConditionalOrderKey(entry.Order.CreatedBy, entry.Order.Nonce)

			if _, exists := store.conditionalOrders[orderKey]; !exists {
				// Already processed by another goroutine
				log.Printf(" Order already processed by another goroutine")
				store.mu.Unlock()
				continue
			}

			// Remove from store before adding to order book to prevent re-triggering
			delete(store.conditionalOrders, orderKey)
			store.mu.Unlock()

			priceToCheckFloat := new(big.Float).Quo(
				new(big.Float).SetInt(priceToCheck),
				new(big.Float).SetInt(PriceFactor),
			)
			triggerPriceFloat := new(big.Float).Quo(
				new(big.Float).SetInt(entry.TriggerPrice),
				new(big.Float).SetInt(PriceFactor),
			)
			ptf, _ := priceToCheckFloat.Float64()
			tpf, _ := triggerPriceFloat.Float64()

			triggerReason = fmt.Sprintf("%s/%s: %.6f %s %.6f (last price)",
				entry.TriggerSymbolIn,
				entry.TriggerSymbolOut,
				ptf,
				map[bool]string{true: ">=", false: "<="}[entry.IsTriggerAbove],
				tpf,
			)

			log.Printf(" Conditional Order TRIGGERED: %s/%s | Type: %s | Reason: %s",
				entry.Order.CreatedBy.Hex()[:10],
				entry.Order.Nonce.String(),
				entry.ConditionalType,
				triggerReason,
			)

			// Add the order to the order book
			log.Printf("🔄 Calling AddOrder...")
			err := store.orderBookStore.AddOrder(entry.Order)
			if err != nil {
				log.Printf(" Error Adding Triggered Conditional Order: %v", err)

				store.mu.Lock()
				store.conditionalOrders[orderKey] = entry
				store.mu.Unlock()

				log.Printf("**Conditional Order Re-queued**: %s/%s (will retry on next check)",
					entry.Order.CreatedBy.Hex()[:10],
					entry.Order.Nonce.String(),
				)
				continue
			}

			log.Printf(" Conditional Order Added to Book: %s/%s | %s %s -> %s %s",
				entry.Order.CreatedBy.Hex()[:10],
				entry.Order.Nonce.String(),
				entry.Order.AmtIn.String(),
				entry.Order.SymbolIn,
				entry.Order.AmtOut.String(),
				entry.Order.SymbolOut,
			)
		} else {
			log.Printf(" Condition NOT met, skipping trigger")
		}
	}
}
