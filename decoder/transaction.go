// Transaction decoder: create, verify, broadcast (signing is done by external MPC)
package decoder

import (
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// TransactionDecoder decodes and submits transactions; each chain implements this interface to plug into the unified build/broadcast flow. Signing is provided by external MPC.
// The entry points for building PendingSignTx are flow.BuildTransaction / flow.BuildSmartContractTransaction; the decoder only builds/verifies/submits rawTx.
// wrapper is WalletDAI and may be nil; when non-nil, implementations may callback to query wallet/account/address/transaction, etc.
type TransactionDecoder interface {
	// CreateRawTransaction builds the raw transaction from rawTx, filling RawHex, Fees, Signatures, etc. When wrapper is non-nil, callbacks may be used (e.g. GetAddress/GetAssetsAccountByAddress, account lists, etc.).
	CreateRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// SignRawTransaction signs rawTx locally and writes back Signatures. May be omitted when signing is done by external MPC.
	SignRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// SubmitRawTransaction submits the signed rawTx to the chain node and returns the transaction result; VerifyRawTransaction should have passed first. When wrapper is non-nil, callbacks may be used.
	SubmitRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) (*types.Transaction, error)
	// VerifyRawTransaction merges and validates rawTx Signatures; when wrapper is non-nil, callbacks may be used (e.g. address ownership checks).
	VerifyRawTransaction(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// GetRawTransactionFeeRate returns the chain's suggested fee rate and unit (e.g. "10", "sat/vB"). When wrapper is non-nil, callbacks may be used.
	GetRawTransactionFeeRate(wrapper wallet.WalletDAI) (feeRate, unit string, err error)
	// EstimateRawTransactionFee estimates fees from rawTx and writes back rawTx.Fees; when wrapper is non-nil, callbacks may be used.
	EstimateRawTransactionFee(wrapper wallet.WalletDAI, rawTx *types.RawTransaction) error
	// CreateSummaryRawTransactionWithError generates multiple RawTransactionWithError from summary params; when wrapper is non-nil, callbacks may be used (e.g. filtering non-business addresses).
	CreateSummaryRawTransactionWithError(wrapper wallet.WalletDAI, sumRawTx *types.SummaryRawTransaction) ([]*types.RawTransactionWithError, error)
}

// TransactionDecoderBase is the transaction decoder base class; unimplemented methods return "not implemented" so chains only need to override required methods.
type TransactionDecoderBase struct{}

// CreateRawTransaction base default returns not implemented.
func (TransactionDecoderBase) CreateRawTransaction(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("CreateRawTransaction")
}

// SignRawTransaction base default returns not implemented (not needed for MPC signing).
func (TransactionDecoderBase) SignRawTransaction(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("SignRawTransaction")
}

// SubmitRawTransaction base default returns not implemented.
func (TransactionDecoderBase) SubmitRawTransaction(wallet.WalletDAI, *types.RawTransaction) (*types.Transaction, error) {
	return nil, errNotImplement("SubmitRawTransaction")
}

// VerifyRawTransaction base default returns not implemented.
func (TransactionDecoderBase) VerifyRawTransaction(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("VerifyRawTransaction")
}

// GetRawTransactionFeeRate base default returns not implemented.
func (TransactionDecoderBase) GetRawTransactionFeeRate(wallet.WalletDAI) (string, string, error) {
	return "", "", errNotImplement("GetRawTransactionFeeRate")
}

// EstimateRawTransactionFee base default returns not implemented.
func (TransactionDecoderBase) EstimateRawTransactionFee(wallet.WalletDAI, *types.RawTransaction) error {
	return errNotImplement("EstimateRawTransactionFee")
}

// CreateSummaryRawTransactionWithError base default returns not implemented.
func (TransactionDecoderBase) CreateSummaryRawTransactionWithError(wallet.WalletDAI, *types.SummaryRawTransaction) ([]*types.RawTransactionWithError, error) {
	return nil, errNotImplement("CreateSummaryRawTransactionWithError")
}

func errNotImplement(method string) error {
	return &types.AdapterError{Code: types.ErrSystemException, Msg: method + " not implement"}
}
