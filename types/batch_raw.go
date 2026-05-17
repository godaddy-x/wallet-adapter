package types

// BatchTransferRecipient 批量中单笔收款方（顺序由具体链适配器定义，须与链上编码一致）。
type BatchTransferRecipient struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}

// BatchRawRequest 批量转账建单入参（跨链统一形态）。
//
// Protocol 为可选扩展字段，由具体链适配器解析与校验；留空时由该适配器按默认批量协议处理。
// ContractAddress、Recipients 等字段语义亦由各链实现约定。
type BatchRawRequest struct {
	Protocol string `json:"protocol,omitempty"`

	AppID string `json:"appId,omitempty"`

	Account *AssetsAccount `json:"account"`
	Coin    Coin           `json:"coin"`

	// ContractAddress 本笔调用的链上合约/程序地址（语义由链适配器定义）；多数场景必填。
	ContractAddress string `json:"contractAddress"`

	Recipients []BatchTransferRecipient `json:"recipients"`

	FeeRate    string            `json:"feeRate,omitempty"`
	ExtParam   map[string]string `json:"extParam,omitempty"`
	Sid        string            `json:"sid,omitempty"`
	TxType     int64             `json:"txType,omitempty"`
	Required   uint64            `json:"reqSigs,omitempty"`
	CreateTime int64             `json:"createTime,omitempty"`

	EstimatedFees    string `json:"estimatedFees,omitempty"`
	EstimatedFeeRate string `json:"estimatedFeeRate,omitempty"`
}
