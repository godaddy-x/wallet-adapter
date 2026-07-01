package types

// BatchTransferRecipient single recipient in a batch transfer (order defined by chain adapter; must match on-chain encoding).
//
//easyjson:json
type BatchTransferRecipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

// BatchRawRequest batch transfer build input (cross-chain unified shape).
//
// Protocol is an optional extension field parsed and validated by the chain adapter; when empty, the adapter uses its default batch protocol.
// ContractAddress, Recipients, and other field semantics are defined by each chain implementation.
//
//easyjson:json
type BatchRawRequest struct {
	Protocol string `json:"protocol,omitempty"`

	AppID string `json:"appId,omitempty"`

	Account *AssetsAccount `json:"account"`
	Coin    Coin           `json:"coin"`

	// ContractAddress on-chain contract/program address for this call (semantics defined by chain adapter); required in most scenarios.
	ContractAddress string `json:"contractAddress"`

	Recipients []BatchTransferRecipient `json:"recipients"`

	FeeRate  string            `json:"feeRate,omitempty"`
	ExtParam map[string]string `json:"extParam,omitempty"`
	// SpeedUp speed up/replace pending batch transaction; same semantics as RawTransaction.SpeedUp.
	SpeedUp    *SpeedUp `json:"speedUp,omitempty"`
	Sid        string   `json:"sid,omitempty"`
	TxType     int64    `json:"txType,omitempty"`
	Required   uint64   `json:"reqSigs,omitempty"`
	CreateTime int64    `json:"createTime,omitempty"`

	EstimatedFees    string `json:"estimatedFees,omitempty"`
	EstimatedFeeRate string `json:"estimatedFeeRate,omitempty"`
}
