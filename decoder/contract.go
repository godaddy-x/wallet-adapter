// 智能合约解析器：代币余额、ABI 调用/创建/广播、合约元数据；与 TransactionDecoder 并列，供链可选实现。
package decoder

import (
	"fmt"

	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// SmartContractDecoder 智能合约解析器接口，供支持合约的链实现（可选）。
// 与 TransactionDecoder 并列；ChainAdapter 通过 GetSmartContractDecoder() 返回，可为 nil。
type SmartContractDecoder interface {
	ABIDAI

	// GetTokenBalanceByAddress 查询地址代币余额列表
	GetTokenBalanceByAddress(contract types.SmartContract, address ...string) ([]*types.TokenBalance, error)
	// CallSmartContractABI 只读调用合约 ABI 方法，不产生链上交易
	CallSmartContractABI(wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) (*types.SmartContractCallResult, *types.AdapterError)
	// CreateSmartContractRawTransaction 创建智能合约原始交易单（填充 Raw 等）
	CreateSmartContractRawTransaction(wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) *types.AdapterError
	// SubmitSmartContractRawTransaction 广播智能合约交易单
	SubmitSmartContractRawTransaction(wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) (*types.SmartContractReceipt, *types.AdapterError)
	// GetTokenMetadata 根据合约地址查询代币元数据
	GetTokenMetadata(contract string) (*types.SmartContract, error)
}

// ABIDAI ABI 数据访问接口，供 SmartContractDecoder 嵌入
type ABIDAI interface {
	GetABIInfo(address string) (*types.ABIInfo, error)
	SetABIInfo(address string, abi types.ABIInfo) error
}

// SmartContractDecoderBase 默认未实现基类
type SmartContractDecoderBase struct{}

func (SmartContractDecoderBase) GetTokenBalanceByAddress(contract types.SmartContract, address ...string) ([]*types.TokenBalance, error) {
	return nil, errNotImplementSC("GetTokenBalanceByAddress")
}

func (SmartContractDecoderBase) CallSmartContractABI(wallet.WalletDAI, *types.SmartContractRawTransaction) (*types.SmartContractCallResult, *types.AdapterError) {
	return nil, types.Errorf(types.ErrSystemException, "CallSmartContractABI not implement")
}

func (SmartContractDecoderBase) CreateSmartContractRawTransaction(wallet.WalletDAI, *types.SmartContractRawTransaction) *types.AdapterError {
	return types.Errorf(types.ErrSystemException, "CreateSmartContractRawTransaction not implement")
}

func (SmartContractDecoderBase) SubmitSmartContractRawTransaction(wallet.WalletDAI, *types.SmartContractRawTransaction) (*types.SmartContractReceipt, *types.AdapterError) {
	return nil, types.Errorf(types.ErrSystemException, "SubmitSmartContractRawTransaction not implement")
}

func (SmartContractDecoderBase) GetABIInfo(address string) (*types.ABIInfo, error) {
	return nil, errNotImplementSC("GetABIInfo")
}

func (SmartContractDecoderBase) SetABIInfo(address string, abi types.ABIInfo) error {
	return errNotImplementSC("SetABIInfo")
}

func (SmartContractDecoderBase) GetTokenMetadata(contract string) (*types.SmartContract, error) {
	return nil, errNotImplementSC("GetTokenMetadata")
}

func errNotImplementSC(method string) error {
	return fmt.Errorf("%s not implement", method)
}
