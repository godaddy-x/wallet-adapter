package types

// TradeOrderOutboundLookupParams queries business trade order outbound snapshot for one on-chain tx.
// One Mongo lookup per txID; filter payer legs in memory when Address is set.
//
//easyjson:json
type TradeOrderOutboundLookupParams struct {
	// TxID on-chain transaction id (required).
	TxID string `json:"txID"`
	// Symbol chain symbol, e.g. BTC (required).
	Symbol string `json:"symbol"`
	// AccountID custodial account id (required).
	AccountID string `json:"accountID"`
	// ContractAddress token contract; empty for native coin.
	ContractAddress string `json:"contractAddress,omitempty"`
	// Address optional payer filter; empty returns all payer legs for the tx.
	Address string `json:"address,omitempty"`
}

// TradeOrderOutboundSnapshot business outbound accounting snapshot for one tx (may contain multiple payers).
//
//easyjson:json
type TradeOrderOutboundSnapshot struct {
	// Found false when no gateway trade record is bound to txID.
	Found bool `json:"found"`
	// Sid business order id from ow_trade broadcast row.
	Sid string `json:"sid"`
	// DataType trade business type (1 normal / 2 summary / 4 batch / …).
	DataType int64 `json:"dataType"`
	// TxID normalized on-chain tx id.
	TxID string `json:"txID"`
	// EstimatedFees whole-tx estimated miner fee from order (optional).
	EstimatedFees string `json:"estimatedFees,omitempty"`
	// Legs per-payer outbound legs parsed from the trade order.
	Legs []TradeOrderPayerLeg `json:"legs"`
}

// TradeOrderPayerLeg one payer address outbound leg from the business order.
//
//easyjson:json
type TradeOrderPayerLeg struct {
	// PayerAddress normalized payer wallet address.
	PayerAddress string `json:"payerAddress"`
	// SendOut total expected external outflow for this payer (coin main unit decimal string).
	SendOut string `json:"sendOut"`
	// TxFromAmount payer vin total recorded at build time (optional cross-check).
	TxFromAmount string `json:"txFromAmount,omitempty"`
	// Outbounds external recipient breakdown (batch / multi-recipient).
	Outbounds []TradeOrderOutboundTarget `json:"outbounds,omitempty"`
}

// TradeOrderOutboundTarget one external recipient amount for a payer leg.
//
//easyjson:json
type TradeOrderOutboundTarget struct {
	Address string `json:"address"`
	Amount  string `json:"amount"`
}
