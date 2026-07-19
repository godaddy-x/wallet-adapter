// Package wallet wallet data access interface (WalletDAI) for TransactionDecoder build/verify callbacks, decoupled from types.
package wallet

import "github.com/godaddy-x/wallet-adapter/types"

// TradeOrderOutboundLookupParams see types.TradeOrderOutboundLookupParams; duplicated here for WalletDAI ergonomics.
type TradeOrderOutboundLookupParams = types.TradeOrderOutboundLookupParams

type SearchParams struct {
	CountQ          bool   // when true, the returned int64 is the total count
	WalletID        string // wallet ID filter
	AccountID       string // account ID filter
	Address         string // address filter
	LastID          int64  // last ID of previous page; 0 means first page
	Limit           int64  // page size
	ContractAddress string // contract address filter
	MinTransfer     string // minimum balance filter
	LastTransfer    string // last recorded balance value
}

// WalletDAI wallet data access interface for flow and TransactionDecoder build/verify callbacks.
// wrapper must not be nil when calling flow.BuildTransaction / BuildSummaryTransaction / SendTransaction (SignPendingTxData or validation is required);
// wrapper may be nil inside decoder callbacks meaning no lookup. Implement only needed methods; WalletDAIBase returns "not implemented" for the rest.
type WalletDAI interface {
	// GetAssetsAccountInfo queries asset account info by account ID. params.AccountID is the account to query.
	GetAssetsAccountInfo(params SearchParams) (*types.AssetsAccount, error)
	// GetAssetsAccountList paginated asset account list. params includes pagination and filters: CountQ whether to return total count (when true, int64 is total); WalletID/AccountID/Address filters; LastID last ID of previous page, 0 for first page; Limit page size.
	GetAssetsAccountList(params SearchParams) ([]*types.AssetsAccount, int64, error)
	// GetAssetsAccountByAddress queries the asset account for an address. params.Address is the address to query.
	GetAssetsAccountByAddress(params SearchParams) (*types.AssetsAccount, error)
	// GetAddress queries address details (public key, labels, etc.). params.Address is the address to query.
	GetAddress(params SearchParams) (*types.Address, error)
	// GetAddressList paginated address list. params includes pagination and filters: CountQ whether to return total count (when true, int64 is total); WalletID/AccountID/Address filters; LastID last ID of previous page, 0 for first page; Limit page size.
	GetAddressList(params SearchParams) ([]*types.Address, int64, error)
	// GetAccountBalanceList queries account balances.
	GetAccountBalanceList(params SearchParams) ([]*types.AssetBalance, int64, error)
	// GetAddressBalanceList queries address balances.
	GetAddressBalanceList(params SearchParams) ([]*types.AssetBalance, int64, error)
	// SetAddressExtParam sets address extension parameter; key is a business-defined key.
	SetAddressExtParam(address string, key string, val interface{}) error
	// GetAddressExtParam gets address extension parameter.
	GetAddressExtParam(address string, key string) (interface{}, error)
	// GetTransactionByTxID queries transaction records by tx ID and chain symbol, for validation, rescan, etc.
	// Returned Transaction uses separated fields:
	//   - FromAddr/FromAmt: sender addresses and amounts (one-to-one)
	//   - ToAddr/ToAmt: receiver addresses and amounts (one-to-one)
	//   - TxAction: direction marker ("send"=outbound, "receive"=inbound, "internal"=internal transfer, "fee"=fee)
	//   - OutputIndex: output index (-2=fee, -1=native coin, 0+=contract event index)
	GetTransactionByTxID(txID, symbol string) ([]*types.Transaction, error)
	// GetTradeOrderOutbound loads business trade order outbound snapshot for one txID (single DB round-trip).
	// params.Address empty: all payer legs; non-empty: filter to that payer in memory.
	// Found=false means no gateway order (not an error).
	GetTradeOrderOutbound(params TradeOrderOutboundLookupParams) (*types.TradeOrderOutboundSnapshot, error)
	// SignPendingTxData signs raw transaction JSON (rawTx) on the business side, returning PendingSignTx with DataSign/TradeSign to ensure transaction payload integrity;
	// SendTransaction calls this again before broadcast to recompute DataSign/TradeSign and verify Data was not tampered with.
	SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error)
}

// WalletDAIBase default not-implemented base for WalletDAI; all methods return "not implemented";
// embed this struct and override only required methods.
type WalletDAIBase struct{}

func (WalletDAIBase) GetAssetsAccountInfo(params SearchParams) (*types.AssetsAccount, error) {
	return nil, errNotImplement("GetAssetsAccountInfo")
}
func (WalletDAIBase) GetAssetsAccountList(params SearchParams) ([]*types.AssetsAccount, int64, error) {
	return nil, 0, errNotImplement("GetAssetsAccountList")
}
func (WalletDAIBase) GetAssetsAccountByAddress(params SearchParams) (*types.AssetsAccount, error) {
	return nil, errNotImplement("GetAssetsAccountByAddress")
}
func (WalletDAIBase) GetAddress(params SearchParams) (*types.Address, error) {
	return nil, errNotImplement("GetAddress")
}
func (WalletDAIBase) GetAddressList(params SearchParams) ([]*types.Address, int64, error) {
	return nil, 0, errNotImplement("GetAddressList")
}
func (WalletDAIBase) GetAccountBalanceList(params SearchParams) ([]*types.AssetBalance, int64, error) {
	return nil, 0, errNotImplement("GetAccountBalanceList")
}
func (WalletDAIBase) GetAddressBalanceList(params SearchParams) ([]*types.AssetBalance, int64, error) {
	return nil, 0, errNotImplement("GetAddressBalanceList")
}
func (WalletDAIBase) SetAddressExtParam(address string, key string, val interface{}) error {
	return errNotImplement("SetAddressExtParam")
}
func (WalletDAIBase) GetAddressExtParam(address string, key string) (interface{}, error) {
	return nil, errNotImplement("GetAddressExtParam")
}
func (WalletDAIBase) GetTransactionByTxID(txID, symbol string) ([]*types.Transaction, error) {
	return nil, errNotImplement("GetTransactionByTxID")
}
func (WalletDAIBase) GetTradeOrderOutbound(params TradeOrderOutboundLookupParams) (*types.TradeOrderOutboundSnapshot, error) {
	return nil, errNotImplement("GetTradeOrderOutbound")
}
func (WalletDAIBase) SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error) {
	return nil, errNotImplement("SignPendingTxData")
}

func errNotImplement(method string) error {
	return &types.AdapterError{Code: types.ErrSystemException, Msg: method + " not implement"}
}
