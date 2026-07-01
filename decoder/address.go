// Address decoder: encode/decode, validation, WIF, multisig, custom address creation.
// Each chain embeds AddressDecoderBase and implements only the methods it needs.
package decoder

import (
	"fmt"

	"github.com/godaddy-x/wallet-adapter/types"
)

// AddressDecoder address decoder interface. Common methods: PublicKeyToAddress, AddressVerify, AddressDecode, AddressEncode; BTC-family chains may implement WIF, RedeemScript; CustomCreateAddress is optional.
type AddressDecoder interface {
	// PrivateKeyToWIF converts a private key to WIF (BTC-family).
	PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error)
	// PublicKeyToAddress converts a public key to an on-chain address.
	PublicKeyToAddress(pub []byte, isTestnet bool) (string, error)
	// WIFToPrivateKey converts WIF to a private key (BTC-family).
	WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error)
	// RedeemScriptToAddress converts a redeem script to a multisig address (BTC-family).
	RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error)
	// AddressDecode decodes an address to an internal representation (e.g. bytes).
	AddressDecode(addr string, opts ...interface{}) ([]byte, error)
	// AddressEncode encodes an internal representation to an on-chain address.
	AddressEncode(pub []byte, opts ...interface{}) (string, error)
	// AddressVerify checks whether the address format is valid.
	AddressVerify(addr string, opts ...interface{}) bool
	// CustomCreateAddress derives a new address per chain rules (optional).
	CustomCreateAddress(account *types.AssetsAccount, newIndex uint64) (*types.Address, error)
	// SupportCustomCreateAddressFunction reports whether CustomCreateAddress is supported.
	SupportCustomCreateAddressFunction() bool
}

// AddressDecoderBase is the address decoder base class; unimplemented methods return "not implemented" or false so chains only implement what they need.
type AddressDecoderBase struct{}

func (AddressDecoderBase) PrivateKeyToWIF([]byte, bool) (string, error) {
	return "", fmt.Errorf("PrivateKeyToWIF not implement")
}
func (AddressDecoderBase) PublicKeyToAddress([]byte, bool) (string, error) {
	return "", fmt.Errorf("PublicKeyToAddress not implement")
}
func (AddressDecoderBase) WIFToPrivateKey(string, bool) ([]byte, error) {
	return nil, fmt.Errorf("WIFToPrivateKey not implement")
}
func (AddressDecoderBase) RedeemScriptToAddress([][]byte, uint64, bool) (string, error) {
	return "", fmt.Errorf("RedeemScriptToAddress not implement")
}
func (AddressDecoderBase) AddressDecode(string, ...interface{}) ([]byte, error) {
	return nil, fmt.Errorf("AddressDecode not implement")
}
func (AddressDecoderBase) AddressEncode([]byte, ...interface{}) (string, error) {
	return "", fmt.Errorf("AddressEncode not implement")
}
func (AddressDecoderBase) AddressVerify(string, ...interface{}) bool {
	return false
}
func (AddressDecoderBase) CustomCreateAddress(*types.AssetsAccount, uint64) (*types.Address, error) {
	return nil, fmt.Errorf("CustomCreateAddress not implement")
}
func (AddressDecoderBase) SupportCustomCreateAddressFunction() bool {
	return false
}
