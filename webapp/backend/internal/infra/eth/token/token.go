package token

import (
	token "dexbe/abi/token"
	"dexbe/internal/infra/eth"
	"log"

	"github.com/ethereum/go-ethereum/common"
)

type TokenContract struct {
	Client *eth.EthClient
	Token  *token.Token
}

func NewTokenContract(ethClient *eth.EthClient, contractAddr string) *TokenContract {
	contractAddress := common.HexToAddress(contractAddr)
	t, err := token.NewToken(contractAddress, ethClient.Client)
	if err != nil {
		log.Fatal(err)
	}
	return &TokenContract{
		Client: ethClient,
		Token:  t,
	}
}

func (contract *TokenContract) GetTokenDetails() map[string]any {
	details := make(map[string]any)
	symbol, err := contract.Token.Symbol(contract.Client.AuthCall)
	if err != nil {
		log.Printf("TOKEN CONTRACT ERROR: %+v", err)
	}
	name, err := contract.Token.Name(contract.Client.AuthCall)
	if err != nil {
		log.Printf("TOKEN CONTRACT ERROR: %+v", err)
	}
	details["name"] = name
	details["symbol"] = symbol
	return details
}
