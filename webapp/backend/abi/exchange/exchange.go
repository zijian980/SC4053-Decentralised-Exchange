// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package exchange

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// ExchangeOrder is an auto generated low-level Go binding around an user-defined struct.
type ExchangeOrder struct {
	CreatedBy common.Address
	SymbolIn  string
	SymbolOut string
	AmtIn     *big.Int
	AmtOut    *big.Int
	Nonce     *big.Int
}

// ExchangeSwapInfo is an auto generated low-level Go binding around an user-defined struct.
type ExchangeSwapInfo struct {
	Order     ExchangeOrder
	Signature []byte
}

// ExchangeMetaData contains all meta data concerning the Exchange contract.
var ExchangeMetaData = &bind.MetaData{
	ABI: "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"registryAddr\",\"type\":\"address\"}],\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"inputs\":[],\"name\":\"ECDSAInvalidSignature\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"uint256\",\"name\":\"length\",\"type\":\"uint256\"}],\"name\":\"ECDSAInvalidSignatureLength\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"bytes32\",\"name\":\"s\",\"type\":\"bytes32\"}],\"name\":\"ECDSAInvalidSignatureS\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"owner\",\"type\":\"address\"}],\"name\":\"OwnableInvalidOwner\",\"type\":\"error\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"account\",\"type\":\"address\"}],\"name\":\"OwnableUnauthorizedAccount\",\"type\":\"error\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"maker\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"taker\",\"type\":\"address\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbolIn\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"string\",\"name\":\"symbolOut\",\"type\":\"string\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amtIn\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"amtOut\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"makerFilledAmtIn\",\"type\":\"uint256\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"takerFilledAmtIn\",\"type\":\"uint256\"}],\"name\":\"OrderExecuted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"internalType\":\"address\",\"name\":\"previousOwner\",\"type\":\"address\"},{\"indexed\":true,\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnershipTransferred\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"address[]\",\"name\":\"makers\",\"type\":\"address[]\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"symbolsIn\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"string[]\",\"name\":\"symbolsOut\",\"type\":\"string[]\"},{\"indexed\":false,\"internalType\":\"uint256[]\",\"name\":\"fillAmounts\",\"type\":\"uint256[]\"},{\"indexed\":false,\"internalType\":\"uint256\",\"name\":\"ringSize\",\"type\":\"uint256\"}],\"name\":\"RingTradeExecuted\",\"type\":\"event\"},{\"inputs\":[],\"name\":\"DOMAIN_SEPARATOR\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"ORDER_TYPEHASH\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"createdBy\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbolIn\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbolOut\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amtIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amtOut\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structExchange.Order\",\"name\":\"order\",\"type\":\"tuple\"}],\"name\":\"computeOrderHash\",\"outputs\":[{\"internalType\":\"bytes32\",\"name\":\"\",\"type\":\"bytes32\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"createdBy\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbolIn\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbolOut\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amtIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amtOut\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structExchange.Order\",\"name\":\"order\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structExchange.SwapInfo\",\"name\":\"makerInfo\",\"type\":\"tuple\"},{\"components\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"createdBy\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbolIn\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbolOut\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amtIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amtOut\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structExchange.Order\",\"name\":\"order\",\"type\":\"tuple\"},{\"internalType\":\"bytes\",\"name\":\"signature\",\"type\":\"bytes\"}],\"internalType\":\"structExchange.SwapInfo\",\"name\":\"takerInfo\",\"type\":\"tuple\"},{\"internalType\":\"uint256\",\"name\":\"fillAmtIn\",\"type\":\"uint256\"}],\"name\":\"executeOrder\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"components\":[{\"internalType\":\"address\",\"name\":\"createdBy\",\"type\":\"address\"},{\"internalType\":\"string\",\"name\":\"symbolIn\",\"type\":\"string\"},{\"internalType\":\"string\",\"name\":\"symbolOut\",\"type\":\"string\"},{\"internalType\":\"uint256\",\"name\":\"amtIn\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"amtOut\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"}],\"internalType\":\"structExchange.Order[]\",\"name\":\"ringOrders\",\"type\":\"tuple[]\"},{\"internalType\":\"uint256[]\",\"name\":\"fillAmounts\",\"type\":\"uint256[]\"}],\"name\":\"executeRingTrade\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"name\":\"filledOrdersAmtIn\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"createdBy\",\"type\":\"address\"},{\"internalType\":\"uint256\",\"name\":\"nonce\",\"type\":\"uint256\"},{\"internalType\":\"uint256\",\"name\":\"originalAmtIn\",\"type\":\"uint256\"}],\"name\":\"getRemainingAmt\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"owner\",\"outputs\":[{\"internalType\":\"address\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"renounceOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"symbolRegistry\",\"outputs\":[{\"internalType\":\"contractITokenRegistry\",\"name\":\"\",\"type\":\"address\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"address\",\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"transferOwnership\",\"outputs\":[],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
}

// ExchangeABI is the input ABI used to generate the binding from.
// Deprecated: Use ExchangeMetaData.ABI instead.
var ExchangeABI = ExchangeMetaData.ABI

// Exchange is an auto generated Go binding around an Ethereum contract.
type Exchange struct {
	ExchangeCaller     // Read-only binding to the contract
	ExchangeTransactor // Write-only binding to the contract
	ExchangeFilterer   // Log filterer for contract events
}

// ExchangeCaller is an auto generated read-only Go binding around an Ethereum contract.
type ExchangeCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExchangeTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ExchangeTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExchangeFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ExchangeFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ExchangeSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ExchangeSession struct {
	Contract     *Exchange         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ExchangeCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ExchangeCallerSession struct {
	Contract *ExchangeCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// ExchangeTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ExchangeTransactorSession struct {
	Contract     *ExchangeTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// ExchangeRaw is an auto generated low-level Go binding around an Ethereum contract.
type ExchangeRaw struct {
	Contract *Exchange // Generic contract binding to access the raw methods on
}

// ExchangeCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ExchangeCallerRaw struct {
	Contract *ExchangeCaller // Generic read-only contract binding to access the raw methods on
}

// ExchangeTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ExchangeTransactorRaw struct {
	Contract *ExchangeTransactor // Generic write-only contract binding to access the raw methods on
}

// NewExchange creates a new instance of Exchange, bound to a specific deployed contract.
func NewExchange(address common.Address, backend bind.ContractBackend) (*Exchange, error) {
	contract, err := bindExchange(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Exchange{ExchangeCaller: ExchangeCaller{contract: contract}, ExchangeTransactor: ExchangeTransactor{contract: contract}, ExchangeFilterer: ExchangeFilterer{contract: contract}}, nil
}

// NewExchangeCaller creates a new read-only instance of Exchange, bound to a specific deployed contract.
func NewExchangeCaller(address common.Address, caller bind.ContractCaller) (*ExchangeCaller, error) {
	contract, err := bindExchange(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ExchangeCaller{contract: contract}, nil
}

// NewExchangeTransactor creates a new write-only instance of Exchange, bound to a specific deployed contract.
func NewExchangeTransactor(address common.Address, transactor bind.ContractTransactor) (*ExchangeTransactor, error) {
	contract, err := bindExchange(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ExchangeTransactor{contract: contract}, nil
}

// NewExchangeFilterer creates a new log filterer instance of Exchange, bound to a specific deployed contract.
func NewExchangeFilterer(address common.Address, filterer bind.ContractFilterer) (*ExchangeFilterer, error) {
	contract, err := bindExchange(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ExchangeFilterer{contract: contract}, nil
}

// bindExchange binds a generic wrapper to an already deployed contract.
func bindExchange(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := ExchangeMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Exchange *ExchangeRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Exchange.Contract.ExchangeCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Exchange *ExchangeRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Exchange.Contract.ExchangeTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Exchange *ExchangeRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Exchange.Contract.ExchangeTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Exchange *ExchangeCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Exchange.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Exchange *ExchangeTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Exchange.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Exchange *ExchangeTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Exchange.Contract.contract.Transact(opts, method, params...)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_Exchange *ExchangeCaller) DOMAINSEPARATOR(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "DOMAIN_SEPARATOR")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_Exchange *ExchangeSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _Exchange.Contract.DOMAINSEPARATOR(&_Exchange.CallOpts)
}

// DOMAINSEPARATOR is a free data retrieval call binding the contract method 0x3644e515.
//
// Solidity: function DOMAIN_SEPARATOR() view returns(bytes32)
func (_Exchange *ExchangeCallerSession) DOMAINSEPARATOR() ([32]byte, error) {
	return _Exchange.Contract.DOMAINSEPARATOR(&_Exchange.CallOpts)
}

// ORDERTYPEHASH is a free data retrieval call binding the contract method 0xf973a209.
//
// Solidity: function ORDER_TYPEHASH() view returns(bytes32)
func (_Exchange *ExchangeCaller) ORDERTYPEHASH(opts *bind.CallOpts) ([32]byte, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "ORDER_TYPEHASH")

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ORDERTYPEHASH is a free data retrieval call binding the contract method 0xf973a209.
//
// Solidity: function ORDER_TYPEHASH() view returns(bytes32)
func (_Exchange *ExchangeSession) ORDERTYPEHASH() ([32]byte, error) {
	return _Exchange.Contract.ORDERTYPEHASH(&_Exchange.CallOpts)
}

// ORDERTYPEHASH is a free data retrieval call binding the contract method 0xf973a209.
//
// Solidity: function ORDER_TYPEHASH() view returns(bytes32)
func (_Exchange *ExchangeCallerSession) ORDERTYPEHASH() ([32]byte, error) {
	return _Exchange.Contract.ORDERTYPEHASH(&_Exchange.CallOpts)
}

// ComputeOrderHash is a free data retrieval call binding the contract method 0x6fba18e2.
//
// Solidity: function computeOrderHash((address,string,string,uint256,uint256,uint256) order) view returns(bytes32)
func (_Exchange *ExchangeCaller) ComputeOrderHash(opts *bind.CallOpts, order ExchangeOrder) ([32]byte, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "computeOrderHash", order)

	if err != nil {
		return *new([32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)

	return out0, err

}

// ComputeOrderHash is a free data retrieval call binding the contract method 0x6fba18e2.
//
// Solidity: function computeOrderHash((address,string,string,uint256,uint256,uint256) order) view returns(bytes32)
func (_Exchange *ExchangeSession) ComputeOrderHash(order ExchangeOrder) ([32]byte, error) {
	return _Exchange.Contract.ComputeOrderHash(&_Exchange.CallOpts, order)
}

// ComputeOrderHash is a free data retrieval call binding the contract method 0x6fba18e2.
//
// Solidity: function computeOrderHash((address,string,string,uint256,uint256,uint256) order) view returns(bytes32)
func (_Exchange *ExchangeCallerSession) ComputeOrderHash(order ExchangeOrder) ([32]byte, error) {
	return _Exchange.Contract.ComputeOrderHash(&_Exchange.CallOpts, order)
}

// FilledOrdersAmtIn is a free data retrieval call binding the contract method 0x3cb2d502.
//
// Solidity: function filledOrdersAmtIn(address , uint256 ) view returns(uint256)
func (_Exchange *ExchangeCaller) FilledOrdersAmtIn(opts *bind.CallOpts, arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "filledOrdersAmtIn", arg0, arg1)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// FilledOrdersAmtIn is a free data retrieval call binding the contract method 0x3cb2d502.
//
// Solidity: function filledOrdersAmtIn(address , uint256 ) view returns(uint256)
func (_Exchange *ExchangeSession) FilledOrdersAmtIn(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Exchange.Contract.FilledOrdersAmtIn(&_Exchange.CallOpts, arg0, arg1)
}

// FilledOrdersAmtIn is a free data retrieval call binding the contract method 0x3cb2d502.
//
// Solidity: function filledOrdersAmtIn(address , uint256 ) view returns(uint256)
func (_Exchange *ExchangeCallerSession) FilledOrdersAmtIn(arg0 common.Address, arg1 *big.Int) (*big.Int, error) {
	return _Exchange.Contract.FilledOrdersAmtIn(&_Exchange.CallOpts, arg0, arg1)
}

// GetRemainingAmt is a free data retrieval call binding the contract method 0x944c33b5.
//
// Solidity: function getRemainingAmt(address createdBy, uint256 nonce, uint256 originalAmtIn) view returns(uint256)
func (_Exchange *ExchangeCaller) GetRemainingAmt(opts *bind.CallOpts, createdBy common.Address, nonce *big.Int, originalAmtIn *big.Int) (*big.Int, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "getRemainingAmt", createdBy, nonce, originalAmtIn)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetRemainingAmt is a free data retrieval call binding the contract method 0x944c33b5.
//
// Solidity: function getRemainingAmt(address createdBy, uint256 nonce, uint256 originalAmtIn) view returns(uint256)
func (_Exchange *ExchangeSession) GetRemainingAmt(createdBy common.Address, nonce *big.Int, originalAmtIn *big.Int) (*big.Int, error) {
	return _Exchange.Contract.GetRemainingAmt(&_Exchange.CallOpts, createdBy, nonce, originalAmtIn)
}

// GetRemainingAmt is a free data retrieval call binding the contract method 0x944c33b5.
//
// Solidity: function getRemainingAmt(address createdBy, uint256 nonce, uint256 originalAmtIn) view returns(uint256)
func (_Exchange *ExchangeCallerSession) GetRemainingAmt(createdBy common.Address, nonce *big.Int, originalAmtIn *big.Int) (*big.Int, error) {
	return _Exchange.Contract.GetRemainingAmt(&_Exchange.CallOpts, createdBy, nonce, originalAmtIn)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Exchange *ExchangeCaller) Owner(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "owner")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Exchange *ExchangeSession) Owner() (common.Address, error) {
	return _Exchange.Contract.Owner(&_Exchange.CallOpts)
}

// Owner is a free data retrieval call binding the contract method 0x8da5cb5b.
//
// Solidity: function owner() view returns(address)
func (_Exchange *ExchangeCallerSession) Owner() (common.Address, error) {
	return _Exchange.Contract.Owner(&_Exchange.CallOpts)
}

// SymbolRegistry is a free data retrieval call binding the contract method 0x9d3e5938.
//
// Solidity: function symbolRegistry() view returns(address)
func (_Exchange *ExchangeCaller) SymbolRegistry(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _Exchange.contract.Call(opts, &out, "symbolRegistry")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// SymbolRegistry is a free data retrieval call binding the contract method 0x9d3e5938.
//
// Solidity: function symbolRegistry() view returns(address)
func (_Exchange *ExchangeSession) SymbolRegistry() (common.Address, error) {
	return _Exchange.Contract.SymbolRegistry(&_Exchange.CallOpts)
}

// SymbolRegistry is a free data retrieval call binding the contract method 0x9d3e5938.
//
// Solidity: function symbolRegistry() view returns(address)
func (_Exchange *ExchangeCallerSession) SymbolRegistry() (common.Address, error) {
	return _Exchange.Contract.SymbolRegistry(&_Exchange.CallOpts)
}

// ExecuteOrder is a paid mutator transaction binding the contract method 0xf0ca0e1a.
//
// Solidity: function executeOrder(((address,string,string,uint256,uint256,uint256),bytes) makerInfo, ((address,string,string,uint256,uint256,uint256),bytes) takerInfo, uint256 fillAmtIn) returns()
func (_Exchange *ExchangeTransactor) ExecuteOrder(opts *bind.TransactOpts, makerInfo ExchangeSwapInfo, takerInfo ExchangeSwapInfo, fillAmtIn *big.Int) (*types.Transaction, error) {
	return _Exchange.contract.Transact(opts, "executeOrder", makerInfo, takerInfo, fillAmtIn)
}

// ExecuteOrder is a paid mutator transaction binding the contract method 0xf0ca0e1a.
//
// Solidity: function executeOrder(((address,string,string,uint256,uint256,uint256),bytes) makerInfo, ((address,string,string,uint256,uint256,uint256),bytes) takerInfo, uint256 fillAmtIn) returns()
func (_Exchange *ExchangeSession) ExecuteOrder(makerInfo ExchangeSwapInfo, takerInfo ExchangeSwapInfo, fillAmtIn *big.Int) (*types.Transaction, error) {
	return _Exchange.Contract.ExecuteOrder(&_Exchange.TransactOpts, makerInfo, takerInfo, fillAmtIn)
}

// ExecuteOrder is a paid mutator transaction binding the contract method 0xf0ca0e1a.
//
// Solidity: function executeOrder(((address,string,string,uint256,uint256,uint256),bytes) makerInfo, ((address,string,string,uint256,uint256,uint256),bytes) takerInfo, uint256 fillAmtIn) returns()
func (_Exchange *ExchangeTransactorSession) ExecuteOrder(makerInfo ExchangeSwapInfo, takerInfo ExchangeSwapInfo, fillAmtIn *big.Int) (*types.Transaction, error) {
	return _Exchange.Contract.ExecuteOrder(&_Exchange.TransactOpts, makerInfo, takerInfo, fillAmtIn)
}

// ExecuteRingTrade is a paid mutator transaction binding the contract method 0x7286df23.
//
// Solidity: function executeRingTrade((address,string,string,uint256,uint256,uint256)[] ringOrders, uint256[] fillAmounts) returns()
func (_Exchange *ExchangeTransactor) ExecuteRingTrade(opts *bind.TransactOpts, ringOrders []ExchangeOrder, fillAmounts []*big.Int) (*types.Transaction, error) {
	return _Exchange.contract.Transact(opts, "executeRingTrade", ringOrders, fillAmounts)
}

// ExecuteRingTrade is a paid mutator transaction binding the contract method 0x7286df23.
//
// Solidity: function executeRingTrade((address,string,string,uint256,uint256,uint256)[] ringOrders, uint256[] fillAmounts) returns()
func (_Exchange *ExchangeSession) ExecuteRingTrade(ringOrders []ExchangeOrder, fillAmounts []*big.Int) (*types.Transaction, error) {
	return _Exchange.Contract.ExecuteRingTrade(&_Exchange.TransactOpts, ringOrders, fillAmounts)
}

// ExecuteRingTrade is a paid mutator transaction binding the contract method 0x7286df23.
//
// Solidity: function executeRingTrade((address,string,string,uint256,uint256,uint256)[] ringOrders, uint256[] fillAmounts) returns()
func (_Exchange *ExchangeTransactorSession) ExecuteRingTrade(ringOrders []ExchangeOrder, fillAmounts []*big.Int) (*types.Transaction, error) {
	return _Exchange.Contract.ExecuteRingTrade(&_Exchange.TransactOpts, ringOrders, fillAmounts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Exchange *ExchangeTransactor) RenounceOwnership(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Exchange.contract.Transact(opts, "renounceOwnership")
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Exchange *ExchangeSession) RenounceOwnership() (*types.Transaction, error) {
	return _Exchange.Contract.RenounceOwnership(&_Exchange.TransactOpts)
}

// RenounceOwnership is a paid mutator transaction binding the contract method 0x715018a6.
//
// Solidity: function renounceOwnership() returns()
func (_Exchange *ExchangeTransactorSession) RenounceOwnership() (*types.Transaction, error) {
	return _Exchange.Contract.RenounceOwnership(&_Exchange.TransactOpts)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Exchange *ExchangeTransactor) TransferOwnership(opts *bind.TransactOpts, newOwner common.Address) (*types.Transaction, error) {
	return _Exchange.contract.Transact(opts, "transferOwnership", newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Exchange *ExchangeSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Exchange.Contract.TransferOwnership(&_Exchange.TransactOpts, newOwner)
}

// TransferOwnership is a paid mutator transaction binding the contract method 0xf2fde38b.
//
// Solidity: function transferOwnership(address newOwner) returns()
func (_Exchange *ExchangeTransactorSession) TransferOwnership(newOwner common.Address) (*types.Transaction, error) {
	return _Exchange.Contract.TransferOwnership(&_Exchange.TransactOpts, newOwner)
}

// ExchangeOrderExecutedIterator is returned from FilterOrderExecuted and is used to iterate over the raw logs and unpacked data for OrderExecuted events raised by the Exchange contract.
type ExchangeOrderExecutedIterator struct {
	Event *ExchangeOrderExecuted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ExchangeOrderExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ExchangeOrderExecuted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ExchangeOrderExecuted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ExchangeOrderExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ExchangeOrderExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ExchangeOrderExecuted represents a OrderExecuted event raised by the Exchange contract.
type ExchangeOrderExecuted struct {
	Maker            common.Address
	Taker            common.Address
	SymbolIn         string
	SymbolOut        string
	AmtIn            *big.Int
	AmtOut           *big.Int
	MakerFilledAmtIn *big.Int
	TakerFilledAmtIn *big.Int
	Raw              types.Log // Blockchain specific contextual infos
}

// FilterOrderExecuted is a free log retrieval operation binding the contract event 0xdee30dc7563c9f9f2d092f5db2c7f7a76a0e16551b7c98febc1f8106d1aa128e.
//
// Solidity: event OrderExecuted(address indexed maker, address indexed taker, string symbolIn, string symbolOut, uint256 amtIn, uint256 amtOut, uint256 makerFilledAmtIn, uint256 takerFilledAmtIn)
func (_Exchange *ExchangeFilterer) FilterOrderExecuted(opts *bind.FilterOpts, maker []common.Address, taker []common.Address) (*ExchangeOrderExecutedIterator, error) {

	var makerRule []interface{}
	for _, makerItem := range maker {
		makerRule = append(makerRule, makerItem)
	}
	var takerRule []interface{}
	for _, takerItem := range taker {
		takerRule = append(takerRule, takerItem)
	}

	logs, sub, err := _Exchange.contract.FilterLogs(opts, "OrderExecuted", makerRule, takerRule)
	if err != nil {
		return nil, err
	}
	return &ExchangeOrderExecutedIterator{contract: _Exchange.contract, event: "OrderExecuted", logs: logs, sub: sub}, nil
}

// WatchOrderExecuted is a free log subscription operation binding the contract event 0xdee30dc7563c9f9f2d092f5db2c7f7a76a0e16551b7c98febc1f8106d1aa128e.
//
// Solidity: event OrderExecuted(address indexed maker, address indexed taker, string symbolIn, string symbolOut, uint256 amtIn, uint256 amtOut, uint256 makerFilledAmtIn, uint256 takerFilledAmtIn)
func (_Exchange *ExchangeFilterer) WatchOrderExecuted(opts *bind.WatchOpts, sink chan<- *ExchangeOrderExecuted, maker []common.Address, taker []common.Address) (event.Subscription, error) {

	var makerRule []interface{}
	for _, makerItem := range maker {
		makerRule = append(makerRule, makerItem)
	}
	var takerRule []interface{}
	for _, takerItem := range taker {
		takerRule = append(takerRule, takerItem)
	}

	logs, sub, err := _Exchange.contract.WatchLogs(opts, "OrderExecuted", makerRule, takerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ExchangeOrderExecuted)
				if err := _Exchange.contract.UnpackLog(event, "OrderExecuted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOrderExecuted is a log parse operation binding the contract event 0xdee30dc7563c9f9f2d092f5db2c7f7a76a0e16551b7c98febc1f8106d1aa128e.
//
// Solidity: event OrderExecuted(address indexed maker, address indexed taker, string symbolIn, string symbolOut, uint256 amtIn, uint256 amtOut, uint256 makerFilledAmtIn, uint256 takerFilledAmtIn)
func (_Exchange *ExchangeFilterer) ParseOrderExecuted(log types.Log) (*ExchangeOrderExecuted, error) {
	event := new(ExchangeOrderExecuted)
	if err := _Exchange.contract.UnpackLog(event, "OrderExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ExchangeOwnershipTransferredIterator is returned from FilterOwnershipTransferred and is used to iterate over the raw logs and unpacked data for OwnershipTransferred events raised by the Exchange contract.
type ExchangeOwnershipTransferredIterator struct {
	Event *ExchangeOwnershipTransferred // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ExchangeOwnershipTransferredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ExchangeOwnershipTransferred)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ExchangeOwnershipTransferred)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ExchangeOwnershipTransferredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ExchangeOwnershipTransferredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ExchangeOwnershipTransferred represents a OwnershipTransferred event raised by the Exchange contract.
type ExchangeOwnershipTransferred struct {
	PreviousOwner common.Address
	NewOwner      common.Address
	Raw           types.Log // Blockchain specific contextual infos
}

// FilterOwnershipTransferred is a free log retrieval operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Exchange *ExchangeFilterer) FilterOwnershipTransferred(opts *bind.FilterOpts, previousOwner []common.Address, newOwner []common.Address) (*ExchangeOwnershipTransferredIterator, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Exchange.contract.FilterLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return &ExchangeOwnershipTransferredIterator{contract: _Exchange.contract, event: "OwnershipTransferred", logs: logs, sub: sub}, nil
}

// WatchOwnershipTransferred is a free log subscription operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Exchange *ExchangeFilterer) WatchOwnershipTransferred(opts *bind.WatchOpts, sink chan<- *ExchangeOwnershipTransferred, previousOwner []common.Address, newOwner []common.Address) (event.Subscription, error) {

	var previousOwnerRule []interface{}
	for _, previousOwnerItem := range previousOwner {
		previousOwnerRule = append(previousOwnerRule, previousOwnerItem)
	}
	var newOwnerRule []interface{}
	for _, newOwnerItem := range newOwner {
		newOwnerRule = append(newOwnerRule, newOwnerItem)
	}

	logs, sub, err := _Exchange.contract.WatchLogs(opts, "OwnershipTransferred", previousOwnerRule, newOwnerRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ExchangeOwnershipTransferred)
				if err := _Exchange.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOwnershipTransferred is a log parse operation binding the contract event 0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0.
//
// Solidity: event OwnershipTransferred(address indexed previousOwner, address indexed newOwner)
func (_Exchange *ExchangeFilterer) ParseOwnershipTransferred(log types.Log) (*ExchangeOwnershipTransferred, error) {
	event := new(ExchangeOwnershipTransferred)
	if err := _Exchange.contract.UnpackLog(event, "OwnershipTransferred", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ExchangeRingTradeExecutedIterator is returned from FilterRingTradeExecuted and is used to iterate over the raw logs and unpacked data for RingTradeExecuted events raised by the Exchange contract.
type ExchangeRingTradeExecutedIterator struct {
	Event *ExchangeRingTradeExecuted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ExchangeRingTradeExecutedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ExchangeRingTradeExecuted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ExchangeRingTradeExecuted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ExchangeRingTradeExecutedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ExchangeRingTradeExecutedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ExchangeRingTradeExecuted represents a RingTradeExecuted event raised by the Exchange contract.
type ExchangeRingTradeExecuted struct {
	Makers      []common.Address
	SymbolsIn   []string
	SymbolsOut  []string
	FillAmounts []*big.Int
	RingSize    *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterRingTradeExecuted is a free log retrieval operation binding the contract event 0x90385986829d7b58d2bb51c1a6cb4d15db99a9321036fed7b9c18d5858d785d2.
//
// Solidity: event RingTradeExecuted(address[] makers, string[] symbolsIn, string[] symbolsOut, uint256[] fillAmounts, uint256 ringSize)
func (_Exchange *ExchangeFilterer) FilterRingTradeExecuted(opts *bind.FilterOpts) (*ExchangeRingTradeExecutedIterator, error) {

	logs, sub, err := _Exchange.contract.FilterLogs(opts, "RingTradeExecuted")
	if err != nil {
		return nil, err
	}
	return &ExchangeRingTradeExecutedIterator{contract: _Exchange.contract, event: "RingTradeExecuted", logs: logs, sub: sub}, nil
}

// WatchRingTradeExecuted is a free log subscription operation binding the contract event 0x90385986829d7b58d2bb51c1a6cb4d15db99a9321036fed7b9c18d5858d785d2.
//
// Solidity: event RingTradeExecuted(address[] makers, string[] symbolsIn, string[] symbolsOut, uint256[] fillAmounts, uint256 ringSize)
func (_Exchange *ExchangeFilterer) WatchRingTradeExecuted(opts *bind.WatchOpts, sink chan<- *ExchangeRingTradeExecuted) (event.Subscription, error) {

	logs, sub, err := _Exchange.contract.WatchLogs(opts, "RingTradeExecuted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ExchangeRingTradeExecuted)
				if err := _Exchange.contract.UnpackLog(event, "RingTradeExecuted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseRingTradeExecuted is a log parse operation binding the contract event 0x90385986829d7b58d2bb51c1a6cb4d15db99a9321036fed7b9c18d5858d785d2.
//
// Solidity: event RingTradeExecuted(address[] makers, string[] symbolsIn, string[] symbolsOut, uint256[] fillAmounts, uint256 ringSize)
func (_Exchange *ExchangeFilterer) ParseRingTradeExecuted(log types.Log) (*ExchangeRingTradeExecuted, error) {
	event := new(ExchangeRingTradeExecuted)
	if err := _Exchange.contract.UnpackLog(event, "RingTradeExecuted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
