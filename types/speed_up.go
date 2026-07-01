package types

// SpeedUp speed up/replace pending transaction parameters.
//
// When present at transaction build time, the chain adapter should use fixed nonce and explicit gas strategy,
// not automatic nonce (e.g. GetAddressNonce) or automatic gas price (e.g. eth_gasPrice + configured bump).
//
// Required: Nonce; gas must specify FeeRate, or BaseFeeRate with FeeBumpWei / FeeBumpPercent.
//
//easyjson:json
type SpeedUp struct {
	// Nonce on-chain transaction sequence (decimal or 0x hex string).
	Nonce string `json:"nonce"`
	// FeeRate fixed gas unit price (chain native coin precision string).
	FeeRate string `json:"feeRate,omitempty"`
	// BaseFeeRate base gas unit price for bumping (e.g. old feeRate); used with FeeBump* when FeeRate is unset.
	BaseFeeRate string `json:"baseFeeRate,omitempty"`
	// FeeBumpWei fixed increment added to base gas price.
	FeeBumpWei string `json:"feeBumpWei,omitempty"`
	// FeeBumpPercent percentage bump on base gas price, e.g. 20 means +20%.
	FeeBumpPercent uint64 `json:"feeBumpPercent,omitempty"`
	// FromAddress payer address (for multi-address accounts, keep consistent with first from).
	FromAddress string `json:"fromAddress,omitempty"`
}

// Active reports whether speed-up mode is enabled.
func (s *SpeedUp) Active() bool {
	return s != nil
}
