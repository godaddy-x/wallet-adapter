// Package types 区块扫描相关数据类型
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// BlockHeader 区块头
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
//easyjson:json
type TxInput struct {
	SourceTxID  string
	SourceIndex uint64
	Recharge    Recharge
}

// TxOutPut 交易输出（到账记录）
//easyjson:json
type TxOutPut struct {
	Recharge Recharge
	ExtParam string `json:"extParam"`
}

// TxExtractData 区块扫描后的交易提取结果
//easyjson:json
type TxExtractData struct {
	TxInputs   []*TxInput   `json:"txInputs"`
	TxOutputs  []*TxOutPut  `json:"txOutputs"`
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
//easyjson:json
type ScanTargetResult struct {
	SourceKey  string
	Exist      bool
	TargetInfo interface{}
}

// BlockchainSyncStatus 链同步状态
//easyjson:json
type BlockchainSyncStatus struct {
	NetworkBlockHeight uint64
	CurrentBlockHeight uint64
	Syncing            bool
}

// SmartContractEvent 合约事件
//easyjson:json
type SmartContractEvent struct {
	Contract *SmartContract `json:"contract"`
	Event    string         `json:"event"`
	Value    string         `json:"value"`
}

// SmartContractReceipt 合约交易回执
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
	ExtParam    string                `json:"extParam"`
}
