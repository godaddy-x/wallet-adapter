// Package types smart contract related data types (used with decoder.SmartContractDecoder)
package types

// TokenBalance token balance (contract + balance info)
//
//easyjson:json
type TokenBalance struct {
	Contract *SmartContract `json:"contract"`
	Balance  *Balance       `json:"balance"`
}

// TxRawType smart contract raw transaction Raw field type
const (
	TxRawTypeHex    = 0 // hex string
	TxRawTypeJSON   = 1 // json string
	TxRawTypeBase64 = 2 // base64 string
)

// SmartContractRawTransaction smart contract raw transaction
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
	// SpeedUp speed up/replace pending contract write transaction; when non-empty, build uses fixed nonce and explicit gas.
	SpeedUp *SpeedUp `json:"speedUp,omitempty"`
	// SignExt signature verification metadata (JSON object string); written only by adapter at build time.
	SignExt string `json:"signExt,omitempty"`
	// Fields below are filled by flow.BuildSmartContractTransaction; same semantics as RawTransaction fields of the same name; TxType=2 means contract write.
	CreateTime  int64  `json:"createTime"`
	CreateNonce string `json:"createNonce"`
	TxType      int64  `json:"txType"`
}

// SmartContractCallResult status constants
const (
	SmartContractCallResultStatusFail    = 0
	SmartContractCallResultStatusSuccess = 1
)

// SmartContractCallResult read-only contract call result (does not produce an on-chain transaction)
//
//easyjson:json
type SmartContractCallResult struct {
	Method    string `json:"method"`
	Value     string `json:"value"`
	RawHex    string `json:"rawHex"`
	Status    uint64 `json:"status"`
	Exception string `json:"exception"`
}

// ABIInfo contract ABI info
//
//easyjson:json
type ABIInfo struct {
	Address string      `json:"address"`
	ABI     interface{} `json:"abi"`
}
