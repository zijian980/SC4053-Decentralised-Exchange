package registry

import (
	"dexbe/abi/registry"
	"dexbe/internal/infra/eth"
	"log"

	"github.com/ethereum/go-ethereum/common"
)

type RegistryContract struct {
	Client   *eth.EthClient
	Registry *registry.Registry
}

func NewRegistryContract(ethClient *eth.EthClient, contractAddr string) *RegistryContract {
	contractAddress := common.HexToAddress(contractAddr)
	ex, err := registry.NewRegistry(contractAddress, ethClient.Client)
	if err != nil {
		log.Fatal(err)
	}
	return &RegistryContract{
		Client:   ethClient,
		Registry: ex,
	}
}

func (contract *RegistryContract) GetAllTokens() []common.Address {
	tokenAddrs, err := contract.Registry.GetAllTokens(contract.Client.AuthCall)
	if err != nil {
		log.Printf("TOKENREGISTRY CONTRACT ERROR: %+v", err)
		return make([]common.Address, 0)
	}
	return tokenAddrs
}

func (contract *RegistryContract) AddToken(name, symbol, address string) {
	addr := common.HexToAddress(address)
	_, err := contract.Registry.AddToken(contract.Client.AuthTransact, addr)
	if err != nil {
		log.Printf("TOKENREGISTRY CONTRACT ADD TOKEN ERROR: %+v", err)
	}
}
