// Package types 智能合约相关数据类型（与 decoder.SmartContractDecoder 配套）
package types

// TokenBalance 代币余额（合约 + 余额信息）
//
//easyjson:json
type TokenBalance struct {
	Contract *SmartContract `json:"contract"`
	Balance  *Balance       `json:"balance"`
}

// TxRawType 智能合约原始交易 Raw 字段类型
const (
	TxRawTypeHex    = 0 // hex 字符串
	TxRawTypeJSON   = 1 // json 字符串
	TxRawTypeBase64 = 2 // base64 字符串
)

// SmartContractRawTransaction 智能合约原始交易单
//
//easyjson:json
type SmartContractRawTransaction struct {
	Coin         Coin                       `json:"coin"`
	TxID         string                     `json:"txID"`
	Sid          string                     `json:"sid"`
	Account      *AssetsAccount             `json:"account"`
	Signatures   map[string][]*KeySignature `json:"signatures"`
	IsBuilt      bool                       `json:"isBuilt"`
	IsCompleted  bool                       `json:"isComplete"`
	IsSubmit     bool                       `json:"isSubmit"`
	Raw          string                     `json:"raw"`
	RawType      uint64                     `json:"rawType"`
	ABIParam     []string                   `json:"abiParam"`
	Value        string                     `json:"value"`
	FeeRate      string                     `json:"feeRate"`
	Fees         string                     `json:"fees"`
	TxFrom       string                     `json:"txFrom"`
	TxTo         string                     `json:"txTo"`
	AwaitResult  bool                       `json:"awaitResult"`
	AwaitTimeout uint64                     `json:"awaitTimeout"`
}

// SmartContractCallResult 状态常量
const (
	SmartContractCallResultStatusFail    = 0
	SmartContractCallResultStatusSuccess = 1
)

// SmartContractCallResult 合约只读调用结果（不产生链上交易）
//
//easyjson:json
type SmartContractCallResult struct {
	Method    string `json:"method"`
	Value     string `json:"value"`
	RawHex    string `json:"rawHex"`
	Status    uint64 `json:"status"`
	Exception string `json:"exception"`
}

// ABIInfo 合约 ABI 信息
//
//easyjson:json
type ABIInfo struct {
	Address string      `json:"address"`
	ABI     interface{} `json:"abi"`
}
