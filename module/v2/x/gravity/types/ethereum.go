package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

const (
	// GravityDenomPrefix indicates the prefix for all assests minted by this module
	GravityDenomPrefix = ModuleName

	// GravityDenomSeparator is the separator for gravity denoms
	GravityDenomSeparator = ""

	// EthereumContractAddressLen is the length of contract address strings
	EthereumContractAddressLen = 42

	// GravityDenomLen is the length of the denoms generated by the gravity module
	GravityDenomLen = len(GravityDenomPrefix) + len(GravityDenomSeparator) + EthereumContractAddressLen
)

// EthAddress Regular EthAddress
type EthAddress struct {
	address common.Address
}

// GetAddress Returns the contained address as a string
func (ea EthAddress) GetAddress() common.Address {
	return ea.address
}

// SetAddress Sets the contained address, performing validation before updating the value
func (ea *EthAddress) SetAddress(address string) error {
	if err := ValidateEthAddress(address); err != nil {
		return err
	}
	ea.address = common.HexToAddress(address)
	return nil
}

func NewEthAddressFromBytes(address []byte) (*EthAddress, error) {
	if err := ValidateEthAddress(hex.EncodeToString(address)); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid input address")
	}

	addr := EthAddress{common.BytesToAddress(address)}
	return &addr, nil
}

// NewEthAddress Creates a new EthAddress from a string, performing validation and returning any validation errors
func NewEthAddress(address string) (*EthAddress, error) {
	if err := ValidateEthAddress(address); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid input address")
	}

	addr := EthAddress{common.HexToAddress(address)}
	return &addr, nil
}

func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

// ValidateEthAddress Validates the input string as an Ethereum Address
// Address must not be empty, have 42 character length, start with 0x and have 40 remaining characters in [0-9a-fA-F]
func ValidateEthAddress(address string) error {
	if address == "" {
		return fmt.Errorf("empty")
	}

	if has0xPrefix(address) {
		address = address[2:]
	}

	// An auditor recommended we should check the error of hex.DecodeString, given that geth's HexToAddress ignores it
	if _, err := hex.DecodeString(address); err != nil {
		return fmt.Errorf("invalid hex with error: %s", err)
	}

	if !common.IsHexAddress(address) {
		return fmt.Errorf("address(%s) doesn't pass format validation", address)
	}

	return nil
}

// ValidateBasic Performs validation on the wrapped string
func (ea EthAddress) ValidateBasic() error {
	return ValidateEthAddress(ea.address.Hex())
}

// EthereumAddrLessThan migrates the Ethereum address less than function
func EthereumAddrLessThan(e, o string) bool {
	return bytes.Compare([]byte(e)[:], []byte(o)[:]) == -1
}

// ValidateEthereumAddress validates the ethereum address strings
// func ValidateEthereumAddress(a string) error {
// 	if a == "" {
// 		return fmt.Errorf("empty")
// 	}
// 	if !regexp.MustCompile("^0x[0-9a-fA-F]{40}$").MatchString(a) {
// 		return fmt.Errorf("address(%s) doesn't pass regex", a)
// 	}
// 	if len(a) != EthereumContractAddressLen {
// 		return fmt.Errorf("address(%s) of the wrong length exp(%d) actual(%d)", a, len(a), EthereumContractAddressLen)
// 	}
// 	return nil
// }

/////////////////////////
//     ERC20Token      //
/////////////////////////

// NewERC20Token returns a new instance of an ERC20
func NewERC20Token(amount uint64, contract common.Address) ERC20Token {
	return ERC20Token{
		Amount:   sdk.NewIntFromUint64(amount),
		Contract: contract.Hex(),
	}
}

func NewSDKIntERC20Token(amount sdk.Int, contract common.Address) ERC20Token {
	return ERC20Token{
		Amount:   amount,
		Contract: contract.Hex(),
	}
}

func GravityDenom(contract common.Address) string {
	return strings.Join([]string{GravityDenomPrefix, contract.Hex()}, GravityDenomSeparator)
}

// GravityCoin returns the gravity representation of the ERC20
func (e ERC20Token) GravityCoin() sdk.Coin {
	return sdk.Coin{Amount: e.Amount, Denom: GravityDenom(common.HexToAddress(e.Contract))}
}

func GravityDenomToERC20(denom string) (string, error) {
	fullPrefix := GravityDenomPrefix + GravityDenomSeparator
	if !strings.HasPrefix(denom, fullPrefix) {
		return "", fmt.Errorf("denom prefix(%s) not equal to expected(%s)", denom, fullPrefix)
	}
	contract := strings.TrimPrefix(denom, fullPrefix)
	switch {
	case !common.IsHexAddress(contract):
		return "", fmt.Errorf("error validating ethereum contract address")
	case len(denom) != GravityDenomLen:
		return "", fmt.Errorf("len(denom)(%d) not equal to GravityDenomLen(%d)", len(denom), GravityDenomLen)
	default:
		return common.HexToAddress(contract).Hex(), nil
	}
}

func NormalizeCoinDenom(coin *sdk.Coin) {
	coin.Denom = NormalizeDenom(coin.Denom)
}

func NormalizeDenom(denom string) string {
	if contract, err := GravityDenomToERC20(denom); err == nil {
		return GravityDenom(common.HexToAddress(contract))
	}

	return denom
}

func NewSendToEthereumTx(id uint64, tokenContract common.Address, sender sdk.AccAddress, recipient common.Address, amount, feeAmount uint64) *SendToEthereum {
	return &SendToEthereum{
		Id:                id,
		Erc20Fee:          NewERC20Token(feeAmount, tokenContract),
		Sender:            sender.String(),
		EthereumRecipient: recipient.Hex(),
		Erc20Token:        NewERC20Token(amount, tokenContract),
	}
}

// Id:                2,
// Erc20Fee:          types.NewERC20Token(3, myTokenContractAddr),
// Sender:            mySender.String(),
// EthereumRecipient: myReceiver,
// Erc20Token:        types.NewERC20Token(101, myTokenContractAddr),
