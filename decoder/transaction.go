// 交易单解析器：创建、验证、广播（签名由外部 MPC 完成）
package decoder

import (
	"github.com/blockchain/wallet-adapter/types"
	"github.com/blockchain/wallet-adapter/wallet"
)

// TransactionDecoder 交易单解析器，每条链实现此接口即可接入统一构建与广播流程；签名由外部 MPC 提供。
// 构建待签 PendingSignTx 的入口为 flow.BuildTransaction，decoder 只负责构建/验签/提交 rawTx。
// wrapper 为 WalletDAI，可为 nil；非 nil 时实现层可回调查询钱包/账户/地址/交易等。
type TransactionDecoder interface {
	// CreateRawTransaction 根据 rawTx 构建原始交易，填充 RawHex、Fees、Signatures 等。若 wrapper 非 nil，可回调查询（如 GetAddress/GetAssetsAccountByAddress、账户列表等）。
	CreateRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// SignRawTransaction 对 rawTx 进行本地签名并写回 Signatures。签名由外部 MPC 完成时可不实现。
	SignRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// SubmitRawTransaction 将已签名的 rawTx 提交到链上节点并返回交易结果；调用前应已通过 VerifyRawTransaction。若 wrapper 非 nil 可回调查询。
	SubmitRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) (*types.Transaction, error)
	// VerifyRawTransaction 合并并校验 rawTx 的 Signatures；若 wrapper 非 nil，可回调查询（如校验地址归属）。
	VerifyRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// GetRawTransactionFeeRate 返回当前链建议的费率与单位（如 "10"、"sat/vB"）。若 wrapper 非 nil 可回调查询。
	GetRawTransactionFeeRate(wrapper wallet.WalletDAI) (feeRate, unit string, err error)
	// EstimateRawTransactionFee 根据 rawTx 估算手续费并写回 rawTx.Fees；若 wrapper 非 nil 可回调查询。
	EstimateRawTransactionFee(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// CreateSummaryRawTransactionWithError 根据汇总参数生成多笔 RawTransactionWithError；若 wrapper 非 nil 可回调查询（如过滤非本业务地址）。
	CreateSummaryRawTransactionWithError(wrapper wallet.WalletDAI, sumRawTx *types.SummaryRawTransaction) ([]*types.RawTransactionWithError, error)
}

// TransactionDecoderBase 交易单解析器基类，未重写的方法均返回“未实现”，便于链实现时只实现必要方法。
type TransactionDecoderBase struct{}

// CreateRawTransaction 基类默认返回未实现。
func (TransactionDecoderBase) CreateRawTransaction(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("CreateRawTransaction")
}

// SignRawTransaction 基类默认返回未实现（MPC 签名场景无需实现）。
func (TransactionDecoderBase) SignRawTransaction(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("SignRawTransaction")
}

// SubmitRawTransaction 基类默认返回未实现。
func (TransactionDecoderBase) SubmitRawTransaction(wallet.WalletDAI, *types.RawTransaction) (*types.Transaction, error) {
	return nil, errNotImplement("SubmitRawTransaction")
}

// VerifyRawTransaction 基类默认返回未实现。
func (TransactionDecoderBase) VerifyRawTransaction(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("VerifyRawTransaction")
}

// GetRawTransactionFeeRate 基类默认返回未实现。
func (TransactionDecoderBase) GetRawTransactionFeeRate(wallet.WalletDAI) (string, string, error) {
	return "", "", errNotImplement("GetRawTransactionFeeRate")
}

// EstimateRawTransactionFee 基类默认返回未实现。
func (TransactionDecoderBase) EstimateRawTransactionFee(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("EstimateRawTransactionFee")
}

// CreateSummaryRawTransactionWithError 基类默认返回未实现。
func (TransactionDecoderBase) CreateSummaryRawTransactionWithError(wallet.WalletDAI, *types.SummaryRawTransaction) ([]*types.RawTransactionWithError, error) {
	return nil, errNotImplement("CreateSummaryRawTransactionWithError")
}

func errNotImplement(method string) error {
	return &types.AdapterError{Code: types.ErrSystemException, Msg: method + " not implement"}
}
