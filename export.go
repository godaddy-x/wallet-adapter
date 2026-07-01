// Unified exports: types, functions from types/wallet/decoder/config/chain/flow/scanner for one-stop import "github.com/godaddy-x/wallet-adapter".
// The config package (Configer, JSON parsing) must be imported separately via "github.com/godaddy-x/wallet-adapter/config" for LoadAssetsConfig and similar reuse.
package adapter

import (
	"math/big"

	"github.com/godaddy-x/wallet-adapter/amount"
	"github.com/godaddy-x/wallet-adapter/chain"
	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/flow"
	"github.com/godaddy-x/wallet-adapter/scanner"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// ----- types exports -----
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
	// block scanning
	BlockHeader            = types.BlockHeader
	UnscanRecord           = types.UnscanRecord
	Balance                = types.Balance
	AssetBalance           = types.AssetBalance
	Recharge               = types.Recharge
	TxInput                = types.TxInput
	TxOutPut               = types.TxOutPut
	TxExtractData          = types.TxExtractData
	ExtractDataItem        = types.ExtractDataItem
	ContractReceiptItem    = types.ContractReceiptItem
	ScanTarget             = types.ScanTarget
	ScanTargetParam        = types.ScanTargetParam
	ScanTargetResult       = types.ScanTargetResult
	BlockchainSyncStatus   = types.BlockchainSyncStatus
	SmartContractReceipt   = types.SmartContractReceipt
	SmartContractEvent     = types.SmartContractEvent
	BatchRawRequest        = types.BatchRawRequest
	BatchTransferRecipient = types.BatchTransferRecipient
	SpeedUp                = types.SpeedUp
	// smart contract decoding
	TokenBalance                = types.TokenBalance
	SmartContractRawTransaction = types.SmartContractRawTransaction
	SmartContractCallResult     = types.SmartContractCallResult
	ABIInfo                     = types.ABIInfo
)

const (
	TxStatusSuccess                  = types.TxStatusSuccess
	TxStatusFail                     = types.TxStatusFail
	BalanceModelTypeAddress          = types.BalanceModelTypeAddress
	BalanceModelTypeAccount          = types.BalanceModelTypeAccount
	ScanTargetTypeAccountAddress     = types.ScanTargetTypeAccountAddress
	ScanTargetTypeAccountAlias       = types.ScanTargetTypeAccountAlias
	ScanTargetTypeContractAddress    = types.ScanTargetTypeContractAddress
	ScanTargetTypeContractAlias      = types.ScanTargetTypeContractAlias
	ScanTargetTypeAddressPubKey      = types.ScanTargetTypeAddressPubKey
	ScanTargetTypeAddressMemo        = types.ScanTargetTypeAddressMemo
	ScanTargetTypeBatchSenderAddress = types.ScanTargetTypeBatchSenderAddress
	// smart contract Raw types
	TxRawTypeHex                         = types.TxRawTypeHex
	TxRawTypeJSON                        = types.TxRawTypeJSON
	TxRawTypeBase64                      = types.TxRawTypeBase64
	SmartContractCallResultStatusFail    = types.SmartContractCallResultStatusFail
	SmartContractCallResultStatusSuccess = types.SmartContractCallResultStatusSuccess
	SmartContractProtocolERC20           = types.SmartContractProtocolERC20
	SmartContractProtocolBatchSender     = types.SmartContractProtocolBatchSender
)

// error codes
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

// ----- decoder exports -----
type TransactionDecoder = decoder.TransactionDecoder
type TransactionDecoderBase = decoder.TransactionDecoderBase
type AddressDecoder = decoder.AddressDecoder
type AddressDecoderBase = decoder.AddressDecoderBase
type SmartContractDecoder = decoder.SmartContractDecoder
type SmartContractDecoderBase = decoder.SmartContractDecoderBase
type ABIDAI = decoder.ABIDAI

// ----- chain exports -----
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

// ----- scanner exports -----
type BlockScanner = scanner.BlockScanner
type BlockScannerBase = scanner.Base
type BlockScanTargetFunc = scanner.BlockScanTargetFunc
type BalanceQueryFunc = scanner.BalanceQueryFunc

var NewBlockScannerBase = scanner.NewBlockScannerBase

// ----- flow exports -----
func BuildTransaction(d TransactionDecoder, wrapper WalletDAI, rawTx *RawTransaction) (*PendingSignTx, error) {
	return flow.BuildTransaction(d, wrapper, rawTx)
}
func BuildSummaryTransaction(d TransactionDecoder, wrapper WalletDAI, sumRawTx *SummaryRawTransaction) ([]*PendingSignTx, error) {
	return flow.BuildSummaryTransaction(d, wrapper, sumRawTx)
}
func SendTransaction(d TransactionDecoder, wrapper WalletDAI, pendingTx *PendingSignTx) (*Transaction, error) {
	return flow.SendTransaction(d, wrapper, pendingTx)
}
func BuildSmartContractTransaction(d SmartContractDecoder, wrapper WalletDAI, rawTx *SmartContractRawTransaction) (*PendingSignTx, error) {
	return flow.BuildSmartContractTransaction(d, wrapper, rawTx)
}
func SendSmartContractTransaction(d SmartContractDecoder, wrapper WalletDAI, pendingTx *PendingSignTx) (*SmartContractReceipt, error) {
	return flow.SendSmartContractTransaction(d, wrapper, pendingTx)
}
func GetRandomSecure(l int) ([]byte, error) { return flow.GetRandomSecure(l) }

// ----- amount exports (prefer subpackage import "github.com/godaddy-x/wallet-adapter/amount"; wrapper functions below are also available) -----

func AmountStringToBigInt(amountStr string, decimal int32) (*big.Int, error) {
	return amount.StringToBigInt(amountStr, int64(decimal))
}

func AmountBigIntToDecimal(b *big.Int, decimal int32) (string, error) {
	return amount.BigIntToDecimal(b, int64(decimal))
}

func AmountHumanToChainUnit(amountStr string, decimal int32) (*big.Int, error) {
	return amount.HumanToChainUnit(amountStr, int64(decimal))
}

func AmountHumanToChainUnitString(amountStr string, decimal int32) (string, error) {
	return amount.HumanToChainUnitString(amountStr, int64(decimal))
}
