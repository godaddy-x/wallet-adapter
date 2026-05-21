package types

// SpeedUp 加速/替换 pending 交易参数。
//
// 创建交易单时若携带本对象，链适配器应使用固定 nonce 与显式 gas 策略，
// 不走自动 nonce（如 GetAddressNonce）与自动 gas 价（如 eth_gasPrice + 配置上浮）。
//
// 必填：Nonce；gas 须指定 FeeRate，或 BaseFeeRate 配合 FeeBumpWei / FeeBumpPercent 之一。
//
//easyjson:json
type SpeedUp struct {
	// Nonce 链上交易序号（十进制或 0x 十六进制字符串）。
	Nonce string `json:"nonce"`
	// FeeRate 固定 gas 单价（链主币精度字符串）。
	FeeRate string `json:"feeRate,omitempty"`
	// BaseFeeRate 加价基准 gas 单价（如旧单 feeRate）；未设 FeeRate 时与 FeeBump* 配合使用。
	BaseFeeRate string `json:"baseFeeRate,omitempty"`
	// FeeBumpWei 在基准 gas 价上增加的固定值。
	FeeBumpWei string `json:"feeBumpWei,omitempty"`
	// FeeBumpPercent 在基准 gas 价上按百分比上浮，如 20 表示 +20%。
	FeeBumpPercent uint64 `json:"feeBumpPercent,omitempty"`
	// FromAddress 指定付款地址（多地址账户时与首笔 from 保持一致）。
	FromAddress string `json:"fromAddress,omitempty"`
}

// Active 是否启用加速模式。
func (s *SpeedUp) Active() bool {
	return s != nil
}
