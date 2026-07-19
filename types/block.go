// Package types block scanning related data types
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// BlockHeader block header
//
//easyjson:json
type BlockHeader struct {
	Hash              string `json:"hash"`
	Confirmations     uint64 `json:"confirmations"`
	Merkleroot        string `json:"merkleroot"`
	Previousblockhash string `json:"previousblockhash"`
	Height            uint64 `json:"height"`
	Version           uint64 `json:"version"`
	Time              uint64 `json:"time"`
	Fork              bool   `json:"fork"`
	Symbol            string `json:"symbol"`
}

// UnscanRecord failed block/transaction scan record
//
//easyjson:json
type UnscanRecord struct {
	ID          string `json:"id"`
	BlockHeight uint64 `json:"blockHeight"`
	TxID        string `json:"txid"`
	Reason      string `json:"reason"`
	Symbol      string `json:"symbol"`
}

// NewUnscanRecord constructs an unscanned record and generates ID
func NewUnscanRecord(height uint64, txID, reason, symbol string) *UnscanRecord {
	plain := fmt.Sprintf("%s_%d_%s", symbol, height, txID)
	h := sha256.Sum256([]byte(plain))
	return &UnscanRecord{
		ID:          hex.EncodeToString(h[:]),
		BlockHeight: height,
		TxID:        txID,
		Reason:      reason,
		Symbol:      symbol,
	}
}

// Balance address balance
//
//easyjson:json
type Balance struct {
	Symbol           string `json:"symbol"`
	AccountID        string `json:"accountID"`
	Address          string `json:"address"`
	ConfirmBalance   string `json:"confirmBalance"`
	UnconfirmBalance string `json:"unconfirmBalance"`
	Balance          string `json:"balance"`
}

// Recharge deposit/inbound record (used by TxInput/TxOutPut)
//
//easyjson:json
type Recharge struct {
	Sid         string `json:"sid"`
	TxID        string `json:"txid"`
	AccountID   string `json:"accountID"`
	Address     string `json:"address"`
	Symbol      string `json:"symbol"`
	Coin        Coin   `json:"coin"`
	Amount      string `json:"amount"`
	Confirm     int64  `json:"confirm"`
	BlockHash   string `json:"blockHash"`
	BlockHeight uint64 `json:"blockHeight"`
	Index       uint64 `json:"index"`
	CreateAt    int64  `json:"createdAt"`
	TxType      uint64 `json:"txType"`
}

// TxInput transaction input (outbound record)
//
//easyjson:json
type TxInput struct {
	SourceTxID  string
	SourceIndex uint64
	Recharge    Recharge
}

// TxOutPut transaction output (inbound record)
//
//easyjson:json
type TxOutPut struct {
	Recharge Recharge
	ExtParam string `json:"extParam"`
}

// TxExtractData transaction extraction result after block scan
//
//easyjson:json
type TxExtractData struct {
	TxInputs    []*TxInput   `json:"txInputs"`
	TxOutputs   []*TxOutPut  `json:"txOutputs"`
	Transaction *Transaction `json:"transaction"`
}

// NewTxExtractData constructs an empty extraction result
func NewTxExtractData() *TxExtractData {
	return &TxExtractData{
		TxInputs:  make([]*TxInput, 0),
		TxOutputs: make([]*TxOutPut, 0),
	}
}

// ScanTarget scan target (V1, deprecated for compatibility)
//
//easyjson:json
type ScanTarget struct {
	Address          string
	PublicKey        string
	Alias            string
	Symbol           string
	BalanceModelType BalanceModelType
}

// Scan target types: ScanTargetParam.ScanTargetType determines ScanTarget map key meaning
const (
	ScanTargetTypeAccountAddress  = 0 // ScanTarget = account address
	ScanTargetTypeAccountAlias    = 1 // ScanTarget = account alias
	ScanTargetTypeContractAddress = 2 // ScanTarget = contract address
	ScanTargetTypeContractAlias   = 3 // ScanTarget = contract alias
	ScanTargetTypeAddressPubKey       = 4 // ScanTarget = address public key
	ScanTargetTypeAddressMemo         = 5 // ScanTarget = address memo
	ScanTargetTypeBatchSenderAddress  = 6 // ScanTarget = our BatchSender contract address (separate from ERC20 token contract whitelist)
)

// ScanTargetParam scan target parameters: ScanTargetType distinguishes whether ScanTarget keys are address / alias / public key / memo, etc.
//
//easyjson:json
type ScanTargetParam struct {
	// ScanTarget batch target map: key is target value (address/alias/public key/memo).
	// value is hit result:
	//   - nil means no hit
	//   - non-nil means hit (address type: accountID string recommended; contract type: *Coin recommended)
	ScanTarget     map[string]interface{}
	Symbol         string
	ScanTargetType uint64 // see ScanTargetType* constants
}

// NewScanTargetParamForAddress constructs scan target params from account address
func NewScanTargetParamForAddress(symbol, address string) ScanTargetParam {
	return ScanTargetParam{
		Symbol:         symbol,
		ScanTarget:     map[string]interface{}{address: nil},
		ScanTargetType: ScanTargetTypeAccountAddress,
	}
}

// NewScanTargetParamForContract constructs scan target params from contract address (ERC20 contract whitelist scenario)
// Business layer can detect contract address lookup via ScanTargetType == ScanTargetTypeContractAddress.
// symbol: chain symbol (e.g. eth)
// contractAddr: ERC20 contract address (stored in ScanTarget field)
func NewScanTargetParamForContract(symbol, contractAddr string) ScanTargetParam {
	return ScanTargetParam{
		Symbol:         symbol,
		ScanTarget:     map[string]interface{}{contractAddr: nil},
		ScanTargetType: ScanTargetTypeContractAddress,
	}
}

// NewScanTargetParamForBatchSender constructs scan target from BatchSender contract address (native coin batchSendNativeToken block-scan crediting).
// Business should query BatchSender from deployment binding table and ERC20 from token contract config; non-nil hit value is sufficient.
func NewScanTargetParamForBatchSender(symbol, batchContractAddr string) ScanTargetParam {
	return ScanTargetParam{
		Symbol:         symbol,
		ScanTarget:     map[string]interface{}{batchContractAddr: nil},
		ScanTargetType: ScanTargetTypeBatchSenderAddress,
	}
}

// NewScanTargetParamForAlias constructs scan target params from account alias
func NewScanTargetParamForAlias(symbol, alias string) ScanTargetParam {
	return ScanTargetParam{
		Symbol:         symbol,
		ScanTarget:     map[string]interface{}{alias: nil},
		ScanTargetType: ScanTargetTypeAccountAlias,
	}
}

// ScanTargetResult scan target result
//
//easyjson:json
type ScanTargetResult struct {
	SourceKey  string
	TargetInfo interface{}
}

// BlockchainSyncStatus chain sync status
//
//easyjson:json
type BlockchainSyncStatus struct {
	NetworkBlockHeight uint64
	CurrentBlockHeight uint64
	Syncing            bool
}

// ExtractDataItem transaction extraction result item aggregated by SourceKey.
//
//easyjson:json
type ExtractDataItem struct {
	SourceKey string           `json:"sourceKey"`
	Data      []*TxExtractData `json:"data"`
}

// ContractReceiptItem contract receipt item.
//
//easyjson:json
type ContractReceiptItem struct {
	Key     string                `json:"key"`
	Receipt *SmartContractReceipt `json:"receipt"`
}

// BlockScanResult single height scan result summary (for external cursor maintenance, retry, and alerts).
// Design goal: scanner does not depend on internal storage (BlockchainDAI optional); external systems decide whether to advance height or rescan from this result.
// Note: chain implementations may fill TxTotal/TxFailed etc. by capability; minimum is Height + Success + ErrorReason.
//
//easyjson:json
type BlockScanResult struct {
	Symbol    string `json:"symbol"`
	Height    uint64 `json:"height"`
	BlockHash string `json:"blockHash"`
	// NetworkBlockHeight chain latest height at scan time (eth_blockNumber, etc.) for progress monitoring and alerts.
	NetworkBlockHeight uint64 `json:"networkBlockHeight"`
	Success            bool   `json:"success"`
	ErrorReason        string `json:"errorReason"`
	TxTotal            uint64 `json:"txTotal"`
	TxFailed           uint64 `json:"txFailed"`
	ExtractedTxs       uint64 `json:"extractedTxs"`
	// FailedTxIDs partial failed tx IDs (truncated) for quick diagnosis; full failure details persisted externally.
	FailedTxIDs []string `json:"failedTxIDs"`
	// FailedTxDetail first extraction failure (txid + cause) when Success=false.
	FailedTxDetail string `json:"failedTxDetail,omitempty"`

	// Header block header for this scan (when block fetch succeeds).
	Header *BlockHeader `json:"header"`
	// ExtractData transaction extraction results aggregated by SourceKey (filled when implementation enables extraction).
	ExtractData []*ExtractDataItem `json:"extractData"`
	// ContractReceipts contract receipt list (key may identify, e.g. contractAddr:logIndex).
	ContractReceipts []*ContractReceiptItem `json:"contractReceipts"`
	// Once marks whether this scan was priority/one-shot (triggered by ScanBlockPrioritize).
	// Business can distinguish: true=priority scan result, false=main RunScanLoop regular scan result.
	Once bool `json:"once"`
}

// TxVerifyResult output of on-chain tx verification by txid returning creditable result set.
// Design goal: external systems re-verify on-chain before crediting (inclusion, success, confirmations) and get extraction results consistent with block scan.
//
//easyjson:json
type TxVerifyResult struct {
	Symbol   string `json:"symbol"`
	TxID     string `json:"txid"`
	Verified bool   `json:"verified"`
	Reason   string `json:"reason"`

	BlockHeight   uint64 `json:"blockHeight"`
	BlockHash     string `json:"blockHash"`
	Confirmations uint64 `json:"confirmations"`
	Status        string `json:"status"`

	// ExtractData transaction extraction results aggregated by SourceKey.
	ExtractData []*ExtractDataItem `json:"extractData"`
	// ContractReceipts contract receipt list (key may identify, e.g. txid:contractAddr:logIndex).
	ContractReceipts []*ContractReceiptItem `json:"contractReceipts"`
}

// TxTransferExpected expected single native/token transfer (for strict pre-credit comparison).
//
//easyjson:json
type TxTransferExpected struct {
	// ContractAddr empty means native transfer; non-empty means contract token (ERC20) transfer.
	ContractAddr string `json:"contractAddr"`
	From         string `json:"from"`
	To           string `json:"to"`
	// Amount decimal string (formatted per decimals), e.g. "1.23".
	Amount string `json:"amount"`
	// Decimals used only when ContractAddr is non-empty; 0 means implementation decides (not recommended; external should specify).
	Decimals uint32 `json:"decimals"`
	// OutputIndex optional: uniquely locates records within the same tx; semantics vary:
	// - EVM contract events: event log index (logIndex)
	// - UTXO chains: transaction output index (vout)
	// <0 means do not use this field.
	OutputIndex int64 `json:"outputIndex"`
}

// TxVerifyExpected expected object for pre-credit verification (external system's pending credit record summary).
//
//easyjson:json
type TxVerifyExpected struct {
	Symbol    string `json:"symbol"`
	TxID      string `json:"txid"`
	BlockHash string `json:"blockHash"`
	Height    uint64 `json:"height"`

	// Transfers expected credit transfers (native and tokens).
	Transfers []*TxTransferExpected `json:"transfers"`
}

// TxVerifyMatchResult return value of VerifyTransactionMatch: on-chain verification + comparison with expected values.
//
//easyjson:json
type TxVerifyMatchResult struct {
	TxID     string `json:"txid"`
	Verified bool   `json:"verified"`
	Reason   string `json:"reason"`

	// Mismatches list of mismatch reasons (for logging/alerts).
	Mismatches []string `json:"mismatches"`

	// Chain on-chain verification and actual extraction results (for external diff diagnosis).
	Chain *TxVerifyResult `json:"chain"`
}

// SmartContractEvent contract event
//
//easyjson:json
type SmartContractEvent struct {
	Contract *SmartContract `json:"contract"`
	Event    string         `json:"event"`
	Value    string         `json:"value"`
}

// SmartContractReceipt contract transaction receipt
//
//easyjson:json
type SmartContractReceipt struct {
	Coin        Coin                  `json:"coin"`
	WxID        string                `json:"wxid"`
	TxID        string                `json:"txid"`
	From        string                `json:"from"`
	To          string                `json:"to"`
	Value       string                `json:"value"`
	Fees        string                `json:"fees"`
	RawReceipt  string                `json:"rawReceipt"`
	Events      []*SmartContractEvent `json:"events"`
	BlockHash   string                `json:"blockHash"`
	BlockHeight uint64                `json:"blockHeight"`
	ConfirmTime int64                 `json:"confirmTime"`
	Status      string                `json:"status"`
	Reason      string                `json:"reason"`
	// OutputIndex output index for this receipt; semantics vary:
	// - EVM chains: event log index (logIndex)
	// - UTXO chains: transaction output index (vout)
	// Used to pinpoint records within the same tx.
	OutputIndex int64 `json:"outputIndex"`
	// ExtParam key-value extension fields for structured extra info (e.g. contract_creation).
	ExtParam map[string]string `json:"extParam"`
}
