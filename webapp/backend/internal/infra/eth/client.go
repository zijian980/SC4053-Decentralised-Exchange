package eth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthClient struct {
	Client       *ethclient.Client
	privateKey   *ecdsa.PrivateKey
	AuthTransact *bind.TransactOpts
	AuthCall     *bind.CallOpts
}

func GetEthClient(chainId, url, privateKeyHex string) *EthClient {
	chainID, _ := stringToBigInt(chainId)
	client, err := ethclient.Dial(url)
	if err != nil {
		log.Fatal(err)
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex[2:])
	if err != nil {
		log.Fatalf("ERROR HexToECDSA: %+v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal(err)
	}
	callOpts := &bind.CallOpts{
		Pending: false,
		Context: context.Background(),
	}
	ethClient := &EthClient{
		Client:       client,
		privateKey:   privateKey,
		AuthTransact: auth,
		AuthCall:     callOpts,
	}
	return ethClient
}

func stringToBigInt(amountStr string) (*big.Int, error) {
	bigIntValue := new(big.Int)
	bigIntValue, success := bigIntValue.SetString(amountStr, 10)
	if !success {
		return nil, fmt.Errorf("invalid number format for big.Int: %s", amountStr)
	}
	return bigIntValue, nil
}

func (c *EthClient) CheckTxReceipt(ctx context.Context, tx *types.Transaction, isReceived chan<- int) {
	receipt, err := bind.WaitMined(ctx, c.Client, tx)
	if err != nil {
		log.Printf("transaction failed during mining: %v", err)
		return
	}

	if receipt.Status != 1 {
		log.Printf("transaction failed on chain (status %d). Hash: %s", receipt.Status, tx.Hash().Hex())
		isReceived <- 0
	} else {
		isReceived <- 1
	}

}
