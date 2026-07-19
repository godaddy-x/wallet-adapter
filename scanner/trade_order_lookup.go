package scanner

import (
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// TradeOrderOutboundQuerier loads business trade order outbound snapshot for block scan accounting.
type TradeOrderOutboundQuerier interface {
	GetTradeOrderOutbound(params wallet.TradeOrderOutboundLookupParams) (*types.TradeOrderOutboundSnapshot, error)
}
