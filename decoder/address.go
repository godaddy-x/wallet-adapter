// 地址解析器：编解码、校验、WIF、多签、自定义创建地址。
// 各链嵌入 AddressDecoderBase 并仅实现需要的方法。
package decoder

import (
	"fmt"

	"github.com/blockchain/wallet-adapter/types"
)

// AddressDecoder 地址解析器接口。通用方法：PublicKeyToAddress、AddressVerify、AddressDecode、AddressEncode；BTC 系可选 WIF、RedeemScript；可选 CustomCreateAddress。
type AddressDecoder interface {
	// PrivateKeyToWIF 将私钥转为 WIF（BTC 系）。
	PrivateKeyToWIF(priv []byte, isTestnet bool) (string, error)
	// PublicKeyToAddress 公钥转链上地址。
	PublicKeyToAddress(pub []byte, isTestnet bool) (string, error)
	// WIFToPrivateKey WIF 转私钥（BTC 系）。
	WIFToPrivateKey(wif string, isTestnet bool) ([]byte, error)
	// RedeemScriptToAddress 赎回脚本转多签地址（BTC 系）。
	RedeemScriptToAddress(pubs [][]byte, required uint64, isTestnet bool) (string, error)
	// AddressDecode 地址解码为内部表示（如字节）。
	AddressDecode(addr string, opts ...interface{}) ([]byte, error)
	// AddressEncode 内部表示编码为链上地址。
	AddressEncode(pub []byte, opts ...interface{}) (string, error)
	// AddressVerify 校验地址格式是否合法。
	AddressVerify(addr string, opts ...interface{}) bool
	// CustomCreateAddress 按链规则派生新地址（可选）。
	CustomCreateAddress(account *types.AssetsAccount, newIndex uint64) (*types.Address, error)
	// SupportCustomCreateAddressFunction 是否支持 CustomCreateAddress。
	SupportCustomCreateAddressFunction() bool
}

// AddressDecoderBase 地址解析器基类，未重写的方法返回“未实现”或 false，便于链只实现需要的方法。
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
