package exchange

import (
	"dexbe/abi/exchange"
	"dexbe/internal/domains/order"
	"dexbe/internal/infra/eth"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type ExchangeContract struct {
	Client   *eth.EthClient
	Exchange *exchange.Exchange
}

func NewExchangeContract(ethClient *eth.EthClient, contractAddr string) *ExchangeContract {
	contractAddress := common.HexToAddress(contractAddr)
	ex, err := exchange.NewExchange(contractAddress, ethClient.Client)
	if err != nil {
		log.Fatal(err)
	}

	return &ExchangeContract{
		Client:   ethClient,
		Exchange: ex,
	}
}

// splitSignature splits a 65-byte signature into v, r, s components
func splitSignature(sig []byte) (v uint8, r [32]byte, s [32]byte, err error) {
	if len(sig) != 65 {
		err = fmt.Errorf("invalid signature length: expected 65, got %d", len(sig))
		return
	}

	copy(r[:], sig[0:32])
	copy(s[:], sig[32:64])
	v = sig[64]

	// Adjust v if necessary (some libraries return 0/1 instead of 27/28)
	if v < 27 {
		v += 27
	}

	if v != 27 && v != 28 {
		err = fmt.Errorf("invalid signature v value: %d", v)
		return
	}

	return
}

func (contract *ExchangeContract) ExecuteMatch(
	makerInfo *order.Order,
	takerInfo *order.Order,
	fillAmtIn *big.Int,
) (*types.Transaction, error) {
	makerOrderABI := exchange.ExchangeOrder{
		CreatedBy: makerInfo.CreatedBy,
		SymbolIn:  makerInfo.SymbolIn,
		SymbolOut: makerInfo.SymbolOut,
		AmtIn:     makerInfo.AmtIn,
		AmtOut:    makerInfo.AmtOut,
		Nonce:     makerInfo.Nonce,
	}

	takerOrderABI := exchange.ExchangeOrder{
		CreatedBy: takerInfo.CreatedBy,
		SymbolIn:  takerInfo.SymbolIn,
		SymbolOut: takerInfo.SymbolOut,
		AmtIn:     takerInfo.AmtIn,
		AmtOut:    takerInfo.AmtOut,
		Nonce:     takerInfo.Nonce,
	}

	makerSwapInfoABI := exchange.ExchangeSwapInfo{
		Order:     makerOrderABI,
		Signature: makerInfo.Signature,
	}

	takerSwapInfoABI := exchange.ExchangeSwapInfo{
		Order:     takerOrderABI,
		Signature: takerInfo.Signature,
	}

	log.Printf("Submitting TX for fillAmt: %s. Maker: %s, Taker: %s",
		fillAmtIn.String(), makerInfo.CreatedBy.Hex()[:10], takerInfo.CreatedBy.Hex()[:10])

	tx, err := contract.Exchange.ExecuteOrder(
		contract.Client.AuthTransact,
		makerSwapInfoABI,
		takerSwapInfoABI,
		fillAmtIn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send executeOrder transaction: %w", err)
	}

	log.Printf("Transaction submitted. Hash: %s", tx.Hash().Hex())
	return tx, nil
}

// ExecuteRingTrade sends the on-chain transaction for a ring match
// No signatures needed since backend validation is trusted
func (contract *ExchangeContract) ExecuteRingTrade(
	ringOrders []*order.Order,
	fillAmounts []*big.Int,
) (*types.Transaction, error) {

	// 1. Prepare order data structures for the contract call (no signatures)
	solRingOrders := make([]exchange.ExchangeOrder, len(ringOrders))

	for i, order := range ringOrders {
		solRingOrders[i] = exchange.ExchangeOrder{
			CreatedBy: order.CreatedBy,
			SymbolIn:  order.SymbolIn,
			SymbolOut: order.SymbolOut,
			AmtIn:     order.AmtIn,
			AmtOut:    order.AmtOut,
			Nonce:     order.Nonce,
		}
	}

	// 2. Log the call
	log.Printf("Submitting Ring Trade TX for %d orders.", len(ringOrders))
	for i, order := range ringOrders {
		log.Printf("  Order %d: %s gives %s wants %s (fill: %s)",
			i+1,
			order.CreatedBy.Hex()[:10],
			order.SymbolIn,
			order.SymbolOut,
			fillAmounts[i].String())
	}

	// 3. Call the smart contract
	tx, err := contract.Exchange.ExecuteRingTrade(
		contract.Client.AuthTransact, // Use the owner's transact opts
		solRingOrders,
		fillAmounts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to send executeRingTrade transaction: %w", err)
	}

	log.Printf("Ring Trade Transaction submitted. Hash: %s", tx.Hash().Hex())
	return tx, nil
}
