// 统一导出：types/wallet/decoder/config/chain/flow/scanner 的类型与函数，便于 import "github.com/godaddy-x/wallet-adapter" 一站式使用。
// config 包（Configer、INI 解析）需单独 import "github.com/godaddy-x/wallet-adapter/config" 使用，供 LoadAssetsConfig 等复用。
package adapter

import (
	"github.com/godaddy-x/wallet-adapter/chain"
	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/flow"
	"github.com/godaddy-x/wallet-adapter/scanner"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// ----- types 导出 -----
type (
	Coin                    = types.Coin
	SmartContract           = types.SmartContract
	PendingSignTx           = types.PendingSignTx
	RawTransaction          = types.RawTransaction
	KeySignature            = types.KeySignature
	Transaction             = types.Transaction
	SummaryRawTransaction   = types.SummaryRawTransaction
	FeesSupportAccount      = types.FeesSupportAccount
	RawTransactionWithError = types.RawTransactionWithError
	AssetsAccount           = types.AssetsAccount
	Address                 = types.Address
	AdapterError            = types.AdapterError
	WalletDAI               = wallet.WalletDAI
	WalletDAIBase           = wallet.WalletDAIBase
	SymbolInfo              = types.SymbolInfo
	SymbolInfoBase          = types.SymbolInfoBase
	BalanceModelType        = types.BalanceModelType
	// 扫块相关
	BlockHeader          = types.BlockHeader
	UnscanRecord         = types.UnscanRecord
	Balance              = types.Balance
	Recharge             = types.Recharge
	TxInput              = types.TxInput
	TxOutPut             = types.TxOutPut
	TxExtractData        = types.TxExtractData
	ExtractDataItem      = types.ExtractDataItem
	ContractReceiptItem  = types.ContractReceiptItem
	ScanTarget           = types.ScanTarget
	ScanTargetParam      = types.ScanTargetParam
	ScanTargetResult     = types.ScanTargetResult
	BlockchainSyncStatus = types.BlockchainSyncStatus
	SmartContractReceipt = types.SmartContractReceipt
	SmartContractEvent   = types.SmartContractEvent
	// 智能合约解析相关
	TokenBalance                  = types.TokenBalance
	SmartContractRawTransaction   = types.SmartContractRawTransaction
	SmartContractCallResult       = types.SmartContractCallResult
	ABIInfo                       = types.ABIInfo
)

const (
	TxStatusSuccess               = types.TxStatusSuccess
	TxStatusFail                  = types.TxStatusFail
	BalanceModelTypeAddress       = types.BalanceModelTypeAddress
	BalanceModelTypeAccount       = types.BalanceModelTypeAccount
	ScanTargetTypeAccountAddress  = types.ScanTargetTypeAccountAddress
	ScanTargetTypeAccountAlias    = types.ScanTargetTypeAccountAlias
	ScanTargetTypeContractAddress = types.ScanTargetTypeContractAddress
	ScanTargetTypeContractAlias   = types.ScanTargetTypeContractAlias
	ScanTargetTypeAddressPubKey   = types.ScanTargetTypeAddressPubKey
	ScanTargetTypeAddressMemo     = types.ScanTargetTypeAddressMemo
	// 智能合约 Raw 类型
	TxRawTypeHex    = types.TxRawTypeHex
	TxRawTypeJSON   = types.TxRawTypeJSON
	TxRawTypeBase64 = types.TxRawTypeBase64
	SmartContractCallResultStatusFail    = types.SmartContractCallResultStatusFail
	SmartContractCallResultStatusSuccess = types.SmartContractCallResultStatusSuccess
)

// 错误码
const (
	ErrInsufficientBalanceOfAccount      = types.ErrInsufficientBalanceOfAccount
	ErrInsufficientBalanceOfAddress      = types.ErrInsufficientBalanceOfAddress
	ErrInsufficientFees                  = types.ErrInsufficientFees
	ErrDustLimit                         = types.ErrDustLimit
	ErrCreateRawTransactionFailed        = types.ErrCreateRawTransactionFailed
	ErrSignRawTransactionFailed          = types.ErrSignRawTransactionFailed
	ErrVerifyRawTransactionFailed        = types.ErrVerifyRawTransactionFailed
	ErrSubmitRawTransactionFailed        = types.ErrSubmitRawTransactionFailed
	ErrInsufficientTokenBalanceOfAddress = types.ErrInsufficientTokenBalanceOfAddress
	ErrAccountNotFound                   = types.ErrAccountNotFound
	ErrAddressNotFound                   = types.ErrAddressNotFound
	ErrContractNotFound                  = types.ErrContractNotFound
	ErrAddressEncodeFailed               = types.ErrAddressEncodeFailed
	ErrAddressDecodeFailed               = types.ErrAddressDecodeFailed
	ErrNonceInvalid                      = types.ErrNonceInvalid
	ErrCallFullNodeAPIFailed             = types.ErrCallFullNodeAPIFailed
	ErrNetworkRequestFailed              = types.ErrNetworkRequestFailed
	ErrUnknownException                  = types.ErrUnknownException
	ErrSystemException                   = types.ErrSystemException
)

func NewError(code uint64, text string) *AdapterError { return types.NewError(code, text) }
func Errorf(code uint64, format string, a ...interface{}) *AdapterError {
	return types.Errorf(code, format, a...)
}
func ConvertError(err error) *AdapterError { return types.ConvertError(err) }
func NewUnscanRecord(height uint64, txID, reason, symbol string) *UnscanRecord {
	return types.NewUnscanRecord(height, txID, reason, symbol)
}
func NewTxExtractData() *TxExtractData { return types.NewTxExtractData() }
func NewScanTargetParamForAddress(symbol, address string) ScanTargetParam {
	return types.NewScanTargetParamForAddress(symbol, address)
}
func NewScanTargetParamForAlias(symbol, alias string) ScanTargetParam {
	return types.NewScanTargetParamForAlias(symbol, alias)
}

// ----- decoder 导出 -----
type TransactionDecoder = decoder.TransactionDecoder
type TransactionDecoderBase = decoder.TransactionDecoderBase
type AddressDecoder = decoder.AddressDecoder
type AddressDecoderBase = decoder.AddressDecoderBase
type SmartContractDecoder    = decoder.SmartContractDecoder
type SmartContractDecoderBase = decoder.SmartContractDecoderBase
type ABIDAI = decoder.ABIDAI

// ----- chain 导出 -----
type ChainAdapter = chain.ChainAdapter
type ChainAdapterBase = chain.ChainAdapterBase
type AssetsConfig = chain.AssetsConfig
type AssetsConfigBase = chain.AssetsConfigBase

func RegAdapter(symbol string, a ChainAdapter)       { chain.RegAdapter(symbol, a) }
func GetAdapter(symbol string) (ChainAdapter, error) { return chain.GetAdapter(symbol) }
func GetTransactionDecoder(symbol string) (TransactionDecoder, error) {
	return chain.GetTransactionDecoder(symbol)
}
func GetBlockScanner(symbol string) (BlockScanner, error)     { return chain.GetBlockScanner(symbol) }
func GetAddressDecoder(symbol string) (AddressDecoder, error) { return chain.GetAddressDecoder(symbol) }
func GetSmartContractDecoder(symbol string) (SmartContractDecoder, error) {
	return chain.GetSmartContractDecoder(symbol)
}
func ListSymbols() []string { return chain.ListSymbols() }

// ----- scanner 导出 -----
type BlockScanner = scanner.BlockScanner
type BlockScannerBase = scanner.Base
type BlockScanTargetFunc = scanner.BlockScanTargetFunc

var NewBlockScannerBase = scanner.NewBlockScannerBase

// ----- flow 导出 -----
func BuildTransaction(d TransactionDecoder, wrapper WalletDAI, rawTx *RawTransaction) (*PendingSignTx, error) {
	return flow.BuildTransaction(d, wrapper, rawTx)
}
func BuildSummaryTransaction(d TransactionDecoder, wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*PendingSignTx, error) {
	return flow.BuildSummaryTransaction(d, wrapper, sumRawTx)
}
func SendTransaction(d TransactionDecoder, wrapper WalletDAI, pendingTx *PendingSignTx) (*Transaction, error) {
	return flow.SendTransaction(d, wrapper, pendingTx)
}
func GetRandomSecure(l int) ([]byte, error) { return flow.GetRandomSecure(l) }
