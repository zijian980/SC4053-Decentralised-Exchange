package registry

import (
	"dexbe/internal/domains/token"
	"dexbe/internal/infra/eth/registry"
	tokenC "dexbe/internal/infra/eth/token"
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

type Registry struct {
	Tokens           map[string]*token.Token
	RegistryContract *registry.RegistryContract
	mu               *sync.RWMutex
}

func NewRegistry(registryContract *registry.RegistryContract) *Registry {
	return &Registry{
		Tokens:           make(map[string]*token.Token),
		RegistryContract: registryContract,
		mu:               &sync.RWMutex{},
	}
}

func (r *Registry) Get(key string) *token.Token {
	value, ok := r.Tokens[key]
	if ok {
		return value
	}
	return nil
}

func (r *Registry) Add(token *token.Token) {
	r.Tokens[token.Symbol] = token
}

func (r *Registry) Remove(symbol string) {
	delete(r.Tokens, symbol)
}

func (r *Registry) Build(tokens *[]common.Address) {
	for i := 0; i < len(*tokens); i++ {
		address := (*tokens)[i]
		tokenContract := tokenC.NewTokenContract(r.RegistryContract.Client, address.Hex())
		details := tokenContract.GetTokenDetails()
		r.Add(token.NewToken(details["name"].(string), details["symbol"].(string), address))
	}
}

func (r *Registry) GetAllSymbols() []string {
	symbols := []string{}
	for _, value := range r.Tokens {
		symbols = append(symbols, value.Symbol)
	}
	return symbols
}
