// Package types 区块扫描相关数据类型
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// BlockHeader 区块头
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

// UnscanRecord 扫描失败的区块/交易记录
//
//easyjson:json
type UnscanRecord struct {
	ID          string `json:"id"`
	BlockHeight uint64 `json:"blockHeight"`
	TxID        string `json:"txid"`
	Reason      string `json:"reason"`
	Symbol      string `json:"symbol"`
}

// NewUnscanRecord 构造未扫记录并生成 ID
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

// Balance 地址余额
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

// Recharge 充值/到账记录（用于 TxInput/TxOutPut）
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

// TxInput 交易输入（出账记录）
//
//easyjson:json
type TxInput struct {
	SourceTxID  string
	SourceIndex uint64
	Recharge    Recharge
}

// TxOutPut 交易输出（到账记录）
//
//easyjson:json
type TxOutPut struct {
	Recharge Recharge
	ExtParam string `json:"extParam"`
}

// TxExtractData 区块扫描后的交易提取结果
//
//easyjson:json
type TxExtractData struct {
	TxInputs    []*TxInput   `json:"txInputs"`
	TxOutputs   []*TxOutPut  `json:"txOutputs"`
	Transaction *Transaction `json:"transaction"`
}

// NewTxExtractData 构造空的提取结果
func NewTxExtractData() *TxExtractData {
	return &TxExtractData{
		TxInputs:  make([]*TxInput, 0),
		TxOutputs: make([]*TxOutPut, 0),
	}
}

// ScanTarget 扫描目标（V1，已废弃兼容用）
//
//easyjson:json
type ScanTarget struct {
	Address          string
	PublicKey        string
	Alias            string
	Symbol           string
	BalanceModelType BalanceModelType
}

// 扫描目标类型：ScanTargetParam.ScanTargetType 决定 ScanTarget 字段的含义
const (
	ScanTargetTypeAccountAddress  = 0 // ScanTarget = 账户地址
	ScanTargetTypeAccountAlias    = 1 // ScanTarget = 账户别名
	ScanTargetTypeContractAddress = 2 // ScanTarget = 合约地址
	ScanTargetTypeContractAlias   = 3 // ScanTarget = 合约别名
	ScanTargetTypeAddressPubKey   = 4 // ScanTarget = 地址公钥
	ScanTargetTypeAddressMemo     = 5 // ScanTarget = 地址备注
)

// ScanTargetParam 扫描目标参数：由 ScanTargetType 区分 ScanTarget 是 address / alias / 公钥 / 备注等
//
//easyjson:json
type ScanTargetParam struct {
	ScanTarget     string // 根据 ScanTargetType 填：地址、别名、公钥或备注
	Symbol         string
	ScanTargetType uint64 // 见 ScanTargetType* 常量
}

// NewScanTargetParamForAddress 用账户地址构造扫描目标参数
func NewScanTargetParamForAddress(symbol, address string) ScanTargetParam {
	return ScanTargetParam{Symbol: symbol, ScanTarget: address, ScanTargetType: ScanTargetTypeAccountAddress}
}

// NewScanTargetParamForAlias 用账户别名构造扫描目标参数
func NewScanTargetParamForAlias(symbol, alias string) ScanTargetParam {
	return ScanTargetParam{Symbol: symbol, ScanTarget: alias, ScanTargetType: ScanTargetTypeAccountAlias}
}

// ScanTargetResult 扫描目标结果
//
//easyjson:json
type ScanTargetResult struct {
	SourceKey  string
	Exist      bool
	TargetInfo interface{}
}

// BlockchainSyncStatus 链同步状态
//
//easyjson:json
type BlockchainSyncStatus struct {
	NetworkBlockHeight uint64
	CurrentBlockHeight uint64
	Syncing            bool
}

// ExtractDataItem 按 SourceKey 聚合的交易提取结果项。
//
//easyjson:json
type ExtractDataItem struct {
	SourceKey string           `json:"sourceKey"`
	Data      []*TxExtractData `json:"data"`
}

// ContractReceiptItem 合约回执项。
//
//easyjson:json
type ContractReceiptItem struct {
	Key     string                `json:"key"`
	Receipt *SmartContractReceipt `json:"receipt"`
}

// BlockScanResult 单次按高度扫块的结果摘要（供外部系统维护游标、重试与告警）。
// 设计目标：扫块器不依赖内部存储（BlockchainDAI 可选），外部系统可依据该结果决定是否推进高度或重扫。
// 注意：不同链实现可按能力填充 TxTotal/TxFailed 等字段；最低要求是 Height + Success + ErrorReason。
//
//easyjson:json
type BlockScanResult struct {
	Symbol    string `json:"symbol"`
	Height    uint64 `json:"height"`
	BlockHash string `json:"blockHash"`
	// NetworkBlockHeight 为本次扫描时刻的链上最新高度（eth_blockNumber 等），便于业务侧观测进度与告警。
	NetworkBlockHeight uint64 `json:"networkBlockHeight"`
	Success            bool   `json:"success"`
	ErrorReason        string `json:"errorReason"`
	TxTotal            uint64 `json:"txTotal"`
	TxFailed           uint64 `json:"txFailed"`
	ExtractedTxs       uint64 `json:"extractedTxs"`
	// FailedTxIDs 记录部分失败交易 ID（截断），用于快速定位问题；完整失败明细由外部系统自行落库。
	FailedTxIDs []string `json:"failedTxIDs"`

	// Header 为本次扫描对应的区块头（若能成功获取区块）。
	Header *BlockHeader `json:"header"`
	// ExtractData 为交易提取结果列表（按 SourceKey 聚合，仅当实现方开启交易提取时填充）。
	ExtractData []*ExtractDataItem `json:"extractData"`
	// ContractReceipts 为合约回执列表（key 字段可用于标识，如 contractAddr:logIndex）。
	ContractReceipts []*ContractReceiptItem `json:"contractReceipts"`
	// Once 标记本次扫描是否为“插队/一次性”扫描（由 ScanBlockPrioritize 触发）。
	// 业务方可据此区分：true=插队扫描结果，false=主线 RunScanLoop 常规扫描结果。
	Once bool `json:"once"`
}

// TxVerifyResult 按 txid 复核链上交易并返回“可入账结果集”的输出。
// 设计目标：外部系统在入账前进行二次链上复核（上链归属、成功状态、确认数），并获取与扫块口径一致的提取结果集。
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

	// ExtractData 为交易提取结果列表（按 SourceKey 聚合）。
	ExtractData []*ExtractDataItem `json:"extractData"`
	// ContractReceipts 为合约回执列表（key 字段可用于标识，如 txid:contractAddr:logIndex）。
	ContractReceipts []*ContractReceiptItem `json:"contractReceipts"`
}

// TxTransferExpected 主币/代币单条转账的期望值（用于入账前严格比对）。
//
//easyjson:json
type TxTransferExpected struct {
	// ContractAddr 为空表示主币转账；非空表示合约代币（ERC20）转账。
	ContractAddr string `json:"contractAddr"`
	From         string `json:"from"`
	To           string `json:"to"`
	// Amount 为十进制字符串（已按 decimals 格式化后的金额），例如 "1.23"。
	Amount string `json:"amount"`
	// Decimals 仅在 ContractAddr 非空时使用；若为 0 表示由实现方自行确定（不推荐，建议外部明确）。
	Decimals uint32 `json:"decimals"`
	// LogIndex 可选：用于唯一定位同一 tx 内的多笔 Transfer 事件；<0 表示不使用。
	LogIndex int64 `json:"logIndex"`
}

// TxVerifyExpected 入账前复核所需的期望对象（外部系统准备入账的记录摘要）。
//
//easyjson:json
type TxVerifyExpected struct {
	Symbol    string `json:"symbol"`
	TxID      string `json:"txid"`
	BlockHash string `json:"blockHash"`
	Height    uint64 `json:"height"`

	// Transfers 为期望入账的转账列表（主币与代币均可）。
	Transfers []*TxTransferExpected `json:"transfers"`
}

// TxVerifyMatchResult VerifyTransactionMatch 的返回值：链上复核 + 与期望值比对的结论。
//
//easyjson:json
type TxVerifyMatchResult struct {
	TxID     string `json:"txid"`
	Verified bool   `json:"verified"`
	Reason   string `json:"reason"`

	// Mismatches 用于返回不一致的原因列表（可直接写日志/告警）。
	Mismatches []string `json:"mismatches"`

	// Chain 为链上复核与实际提取结果（便于外部定位差异）。
	Chain *TxVerifyResult `json:"chain"`
}

// SmartContractEvent 合约事件
//
//easyjson:json
type SmartContractEvent struct {
	Contract *SmartContract `json:"contract"`
	Event    string         `json:"event"`
	Value    string         `json:"value"`
}

// SmartContractReceipt 合约交易回执
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
	// LogIndex 表示该回执对应的链上事件序号（用于精确定位同一 tx 内多笔事件）。
	// 普通不含事件序号的场景保持默认值 0。
	LogIndex int64 `json:"logIndex"`
	// ExtParam 为键值对扩展字段，便于结构化存储额外信息（如 contract_creation 等）。
	ExtParam map[string]string `json:"extParam"`
}
