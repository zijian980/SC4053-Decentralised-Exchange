package order

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// Domain Hashing components
const EIP712DomainName = "SmashDEX"
const EIP712DomainVersion = "1"

// Type String for the Domain
const EIP712DomainType = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"

// Type String for the Order
const EIP712OrderType = "Order(address createdBy,string symbolIn,string symbolOut,uint256 amtIn,uint256 amtOut,uint256 nonce)"

type OrderStatus int

const (
	Matching OrderStatus = iota
	PendingConfirmation
	Completed
	Cancelled
	PartialFill
)

type Order struct {
	CreatedBy         common.Address `json:"createdBy"`
	SymbolIn          string         `json:"symbolIn"`
	SymbolOut         string         `json:"symbolOut"`
	AmtIn             *big.Int       `json:"amtIn"`
	AmtOut            *big.Int       `json:"amtOut"`
	Nonce             *big.Int       `json:"nonce"`
	Signature         []byte         `json:"signature"`
	LimitPrice        *big.Int       `json:"limitPrice,omitempty"`
	TriggerPrice      *big.Int       `json:"triggerPrice,omitempty"`
	FilledAmtIn       *big.Int       `json:"filledAmtIn,omitempty"`
	Status            OrderStatus    `json:"status"`
	ConditionalOrder  *Order         `json:"conditionalOrder"`
	TransactionHashes []string
}

func NewOrder(createdBy string, symbolIn string, symbolOut string, amtIn string, amtOut string, nonce string, signature string, limitPrice string, filledAmtIn string, status int, conditionalOrder *Order, triggerPrice string) *Order {
	bigIntAmtIn, _ := stringToBigInt(amtIn)
	bigIntAmtOut, _ := stringToBigInt(amtOut)
	bigIntNonce, _ := stringToBigInt(nonce)
	bigIntFilledAmtIn, _ := stringToBigInt(filledAmtIn)
	decoded_sign, _ := hexutil.Decode(signature)
	var bigIntLimitPrice *big.Int
	if limitPrice != "" {
		bigIntLimitPrice, _ = stringToBigInt(limitPrice)
	} else {
		bigIntLimitPrice = nil
	}
	var bigIntTriggerPrice *big.Int
	if limitPrice != "" {
		bigIntTriggerPrice, _ = stringToBigInt(triggerPrice)
	} else {
		bigIntTriggerPrice = nil
	}
	return &Order{
		CreatedBy:        common.HexToAddress(createdBy),
		SymbolIn:         symbolIn,
		SymbolOut:        symbolOut,
		AmtIn:            bigIntAmtIn,
		AmtOut:           bigIntAmtOut,
		Nonce:            bigIntNonce,
		Signature:        decoded_sign,
		FilledAmtIn:      bigIntFilledAmtIn,
		LimitPrice:       bigIntLimitPrice,
		Status:           OrderStatus(status),
		TriggerPrice:     bigIntTriggerPrice,
		ConditionalOrder: conditionalOrder,
	}
}

func NewOrderOriginal(createdBy common.Address, symbolIn string, symbolOut string, amtIn *big.Int, amtOut *big.Int, nonce *big.Int) *Order {
	return &Order{
		CreatedBy: createdBy,
		SymbolIn:  symbolIn,
		SymbolOut: symbolOut,
		AmtIn:     amtIn,
		AmtOut:    amtOut,
		Nonce:     nonce,
	}
}

// DeepCopy creates a deep copy of the order
func (o *Order) DeepCopy() *Order {
	if o == nil {
		return nil
	}

	orderCopy := &Order{
		CreatedBy: o.CreatedBy, // common.Address is a fixed-size array, so it's copied by value
		SymbolIn:  o.SymbolIn,  // strings are immutable in Go
		SymbolOut: o.SymbolOut,
		Status:    o.Status,
	}

	// Deep copy big.Int pointers
	if o.AmtIn != nil {
		orderCopy.AmtIn = new(big.Int).Set(o.AmtIn)
	}
	if o.AmtOut != nil {
		orderCopy.AmtOut = new(big.Int).Set(o.AmtOut)
	}
	if o.Nonce != nil {
		orderCopy.Nonce = new(big.Int).Set(o.Nonce)
	}
	if o.LimitPrice != nil {
		orderCopy.LimitPrice = new(big.Int).Set(o.LimitPrice)
	}
	if o.FilledAmtIn != nil {
		orderCopy.FilledAmtIn = new(big.Int).Set(o.FilledAmtIn)
	}

	// Deep copy byte slice
	if o.Signature != nil {
		orderCopy.Signature = make([]byte, len(o.Signature))
		copy(orderCopy.Signature, o.Signature)
	}

	// Deep copy string slice
	if o.TransactionHashes != nil {
		orderCopy.TransactionHashes = make([]string, len(o.TransactionHashes))
		copy(orderCopy.TransactionHashes, o.TransactionHashes)
	}

	// Recursively deep copy conditional order
	if o.ConditionalOrder != nil {
		orderCopy.ConditionalOrder = o.ConditionalOrder.DeepCopy()
	}

	return orderCopy
}

func (o *Order) ToStringMap() map[string]string {
	result := make(map[string]string)
	v := reflect.ValueOf(o).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		if !field.CanInterface() {
			continue
		}

		key := fieldType.Name
		val := field.Interface()

		for reflect.TypeOf(val).Kind() == reflect.Ptr {
			rv := reflect.ValueOf(val)
			if rv.IsNil() {
				val = ""
				break
			}
			val = rv.Elem().Interface()
		}

		var strVal string

		switch f := val.(type) {
		case string:
			strVal = f
		case *big.Int:
			if f != nil {
				strVal = f.String()
			} else {
				strVal = ""
			}
		case big.Int:
			strVal = f.String()
		case common.Address:
			strVal = f.Hex()
		case []byte:
			if f != nil {
				strVal = hex.EncodeToString(f)
			}
		case []string:
			strVal = strings.Join(f, ",")
		case OrderStatus:
			switch f {
			case Matching:
				strVal = "Matching"
			case PendingConfirmation:
				strVal = "PendingConfirmation"
			case Completed:
				strVal = "Completed"
			case Cancelled:
				strVal = "Cancelled"
			case PartialFill:
				strVal = "PartialFill"
			default:
				strVal = "Unknown"
			}
		default:
			rv := reflect.ValueOf(f)
			if rv.Kind() == reflect.Struct {
				// Handle nested Order struct properly
				if nestedOrder, ok := val.(Order); ok {
					nestedMap := nestedOrder.ToStringMap()

					// Marshal without HTML escaping
					buf := new(bytes.Buffer)
					enc := json.NewEncoder(buf)
					enc.SetEscapeHTML(false)
					_ = enc.Encode(nestedMap)

					strVal = strings.TrimSpace(buf.String())
				} else {
					buf := new(bytes.Buffer)
					enc := json.NewEncoder(buf)
					enc.SetEscapeHTML(false)
					_ = enc.Encode(f)

					strVal = strings.TrimSpace(buf.String())
				}
			} else {
				strVal = fmt.Sprintf("%v", f)
			}
		}

		result[key] = strVal
	}

	return result
}

func stringToBigInt(amountStr string) (*big.Int, error) {
	bigIntValue := new(big.Int)
	bigIntValue, success := bigIntValue.SetString(amountStr, 10)
	if !success {
		return nil, fmt.Errorf("invalid number format for big.Int: %s", amountStr)
	}
	return bigIntValue, nil
}

func (o *Order) HashOrder(chainID *big.Int, verifyingContract common.Address) ([]byte, error) {
	// Hash types of order
	orderTypeHash := crypto.Keccak256([]byte(EIP712OrderType))
	// Hash dynamic types
	symbolInSlice := crypto.Keccak256([]byte(o.SymbolIn))
	symbolOutSlice := crypto.Keccak256([]byte(o.SymbolOut))

	var symbolInHash [32]byte
	copy(symbolInHash[:], symbolInSlice)

	var symbolOutHash [32]byte
	copy(symbolOutHash[:], symbolOutSlice)
	addrType, err := abi.NewType("address", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create address type: %w", err)
	}
	bytes32Type, err := abi.NewType("bytes32", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create bytes32 type: %w", err)
	}
	uint256Type, err := abi.NewType("uint256", "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create uint256 type: %w", err)
	}

	orderArgs := abi.Arguments{
		{Type: addrType},
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: uint256Type},
		{Type: uint256Type},
	}

	packedOrderData, err := orderArgs.Pack(
		o.CreatedBy,
		symbolInHash,
		symbolOutHash,
		o.AmtIn,
		o.AmtOut,
		o.Nonce,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to EIP-712 encode Order struct data: %w", err)
	}

	structHashInput := append(orderTypeHash, packedOrderData...)
	structHash := crypto.Keccak256(structHashInput)

	domainTypeHash := crypto.Keccak256([]byte(EIP712DomainType))
	nameHashSlice := crypto.Keccak256([]byte(EIP712DomainName))
	versionHashSlice := crypto.Keccak256([]byte(EIP712DomainVersion))
	var nameHash [32]byte
	copy(nameHash[:], nameHashSlice)

	var versionHash [32]byte
	copy(versionHash[:], versionHashSlice)

	domainArgs := abi.Arguments{
		{Type: bytes32Type},
		{Type: bytes32Type},
		{Type: uint256Type},
		{Type: addrType},
	}
	packedDomainData, err := domainArgs.Pack(
		nameHash,
		versionHash,
		chainID,
		verifyingContract,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to EIP-712 encode domain data: %w", err)
	}
	domainHashInput := append(domainTypeHash, packedDomainData...)
	domainSeparator := crypto.Keccak256(domainHashInput)
	prefix := []byte{0x19, 0x01}

	finalHashInput := append(prefix, domainSeparator...)
	finalHashInput = append(finalHashInput, structHash...)

	return crypto.Keccak256(finalHashInput), nil
}

func (o *Order) VerifyOrder(sig []byte, chainID *big.Int, verifyingContract common.Address) (bool, error) {
	if len(sig) != 65 {
		return false, fmt.Errorf("invalid signature length: got %d, want 65", len(sig))
	}

	hashedOrder, err := o.HashOrder(chainID, verifyingContract)
	if err != nil {
		return false, fmt.Errorf("failed to hash order using EIP-712: %w", err)
	}

	v := int(sig[64])
	if v == 27 || v == 28 {
		sig[64] = byte(v - 27)
	}

	pubKey, err := crypto.SigToPub(hashedOrder, sig)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key from signature: %w", err)
	}

	address := crypto.PubkeyToAddress(*pubKey)
	log.Printf("RECOVERED ADDRESS: %+v", address)
	if address != o.CreatedBy {
		return false, fmt.Errorf("signature recovered address %s does not match creator %s", address.Hex(), o.CreatedBy.Hex())
	}

	return true, nil
}
