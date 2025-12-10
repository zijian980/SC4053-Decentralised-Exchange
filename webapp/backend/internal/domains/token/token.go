package token

import "github.com/ethereum/go-ethereum/common"

type Token struct {
	Name    string         `json:"name"`
	Symbol  string         `json:"symbol"`
	Address common.Address `json:"address"`
}

func NewToken(name, symbol string, address common.Address) *Token {
	return &Token{
		Name:    name,
		Symbol:  symbol,
		Address: address,
	}
}
