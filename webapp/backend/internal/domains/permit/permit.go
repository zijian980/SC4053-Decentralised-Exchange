package permit

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"math/big"
)

// Domain Hashing components
const EIP712DomainVersion = "1"

// Type String for the Domain
const EIP712DomainType = "EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"
const EIP712PermitType = "Permit(address owner,address spender,uint256 value,uint256 nonce,uint256 deadline)"

type Permit struct {
	Owner    common.Address
	Spender  common.Address
	Value    *big.Int
	Nonce    *big.Int
	Deadline *big.Int
}

func NewPermit(owner, spender, value, nonce, deadline string) *Permit {
	bigIntValue, _ := stringToBigInt(value)
	bigIntNonce, _ := stringToBigInt(nonce)
	bigIntDeadline, _ := stringToBigInt(deadline)
	return &Permit{
		Owner:    common.HexToAddress(owner),
		Spender:  common.HexToAddress(spender),
		Value:    bigIntValue,
		Nonce:    bigIntNonce,
		Deadline: bigIntDeadline,
	}
}

func stringToBigInt(amountStr string) (*big.Int, error) {
	bigIntValue := new(big.Int)
	bigIntValue, success := bigIntValue.SetString(amountStr, 10)
	if !success {
		return nil, fmt.Errorf("invalid number format for big.Int: %s", amountStr)
	}
	return bigIntValue, nil
}

func (permitData *Permit) HashPermit(tokenAddress common.Address, chainID *big.Int, coinName string) ([]byte, error) {
	permitTypeHash := crypto.Keccak256([]byte(EIP712PermitType))

	addrType, _ := abi.NewType("address", "", nil)
	uint256Type, _ := abi.NewType("uint256", "", nil)

	permitArgs := abi.Arguments{
		{Type: addrType}, {Type: addrType}, {Type: uint256Type},
		{Type: uint256Type}, {Type: uint256Type},
	}

	packedPermitData, err := permitArgs.Pack(
		permitData.Owner,
		permitData.Spender,
		permitData.Value,
		permitData.Nonce,
		permitData.Deadline,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to ABI pack Permit data: %w", err)
	}

	structHash := crypto.Keccak256(append(permitTypeHash, packedPermitData...))

	bytes32Type, _ := abi.NewType("bytes32", "", nil)

	domainTypeHash := crypto.Keccak256([]byte(EIP712DomainType))
	nameHashSlice := crypto.Keccak256([]byte(coinName))
	versionHashSlice := crypto.Keccak256([]byte("1"))

	var nameHash [32]byte
	copy(nameHash[:], nameHashSlice)
	var versionHash [32]byte
	copy(versionHash[:], versionHashSlice)

	domainArgs := abi.Arguments{
		{Type: bytes32Type}, {Type: bytes32Type}, {Type: uint256Type}, {Type: addrType},
	}

	packedDomainData, err := domainArgs.Pack(
		nameHash,
		versionHash,
		chainID,
		tokenAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to ABI pack Permit domain data: %w", err)
	}
	domainSeparator := crypto.Keccak256(append(domainTypeHash, packedDomainData...))

	prefix := []byte{0x19, 0x01}
	finalHashInput := append(prefix, domainSeparator...)
	finalHashInput = append(finalHashInput, structHash...)

	return crypto.Keccak256(finalHashInput), nil
}

func (permit *Permit) VerifyPermit(sig []byte, chainID *big.Int, tokenAddress common.Address, tokenName string) (bool, error) {

	hashedPermit, err := permit.HashPermit(tokenAddress, chainID, tokenName)
	if err != nil {
		return false, fmt.Errorf("failed to hash permit message: %w", err)
	}

	if len(sig) != 65 {
		return false, fmt.Errorf("invalid signature length: got %d, want 65", len(sig))
	}

	v := int(sig[64])
	if v == 27 || v == 28 {
		sig[64] = byte(v - 27)
	}

	pubKey, err := crypto.SigToPub(hashedPermit, sig)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key from permit signature: %w", err)
	}

	recoveredAddress := crypto.PubkeyToAddress(*pubKey)

	if recoveredAddress != permit.Owner {
		return false, fmt.Errorf("permit signature recovered address %s does not match expected owner %s", recoveredAddress.Hex(), permit.Owner.Hex())
	}

	return true, nil
}
