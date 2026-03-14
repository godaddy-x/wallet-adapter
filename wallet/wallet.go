// Package wallet 钱包数据访问接口（WalletDAI），供 TransactionDecoder 在构建/验证时回调查询，与 types 解耦。
package wallet

import "github.com/blockchain/wallet-adapter/types"

// WalletDAI 钱包数据访问接口，供 flow 与 TransactionDecoder 在构建/验证时回调查询。
// 调用 flow.BuildTransaction / BuildSummaryTransaction / SendTransaction 时 wrapper 不可为 nil（需调用 SignPendingTxData 或校验）；
// decoder 内部回调查询时 wrapper 可为 nil 表示不回查。实现方仅实现需要的方法即可，未实现由 WalletDAIBase 返回“未实现”。
type WalletDAI interface {
	// GetAssetsAccountInfo 根据账户 ID 查询资产账户信息。
	GetAssetsAccountInfo(accountID string) (*types.AssetsAccount, error)
	// GetAssetsAccountList 分页查询资产账户列表，lastID 为上一页最后一条 ID，0 表示首页；cols 可选指定返回列。
	GetAssetsAccountList(lastID int64, cols ...interface{}) ([]*types.AssetsAccount, error)
	// GetAssetsAccountByAddress 根据地址查询所属资产账户。
	GetAssetsAccountByAddress(address string) (*types.AssetsAccount, error)
	// GetAddress 根据地址查询地址详情（含公钥、标签等）。
	GetAddress(address string) (*types.Address, error)
	// GetAddressList 分页查询地址列表，lastID 为上一页最后一条 ID，0 表示首页；cols 可选指定返回列。
	GetAddressList(lastID int64, cols ...interface{}) ([]*types.Address, error)
	// SetAddressExtParam 设置地址扩展参数，key 为业务自定义键。
	SetAddressExtParam(address string, key string, val interface{}) error
	// GetAddressExtParam 获取地址扩展参数。
	GetAddressExtParam(address string, key string) (interface{}, error)
	// GetTransactionByTxID 根据交易 ID 与链标识查询交易记录，用于校验、补扫等。
	GetTransactionByTxID(txID, symbol string) ([]*types.Transaction, error)
	// SignPendingTxData 对原始交易 JSON（rawTx）做业务侧签名，返回填充了 DataSign/TradeSign 的 PendingSignTx，用于保证创建交易单数据有效性；
	// SendTransaction 广播前会再次调用本方法复算 DataSign/TradeSign，用于校验 Data 未被篡改。
	SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error)
}

// WalletDAIBase 为 WalletDAI 的默认未实现基类，所有方法均返回“未实现”错误；
// 实现方嵌入此结构体并仅重写需要的方法即可。
type WalletDAIBase struct{}

func (WalletDAIBase) GetAssetsAccountInfo(accountID string) (*types.AssetsAccount, error) {
	return nil, errNotImplement("GetAssetsAccountInfo")
}
func (WalletDAIBase) GetAssetsAccountList(lastID int64, cols ...interface{}) ([]*types.AssetsAccount, error) {
	return nil, errNotImplement("GetAssetsAccountList")
}
func (WalletDAIBase) GetAssetsAccountByAddress(address string) (*types.AssetsAccount, error) {
	return nil, errNotImplement("GetAssetsAccountByAddress")
}
func (WalletDAIBase) GetAddress(address string) (*types.Address, error) {
	return nil, errNotImplement("GetAddress")
}
func (WalletDAIBase) GetAddressList(lastID int64, cols ...interface{}) ([]*types.Address, error) {
	return nil, errNotImplement("GetAddressList")
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
func (WalletDAIBase) SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error) {
	return nil, errNotImplement("SignPendingTxData")
}

func errNotImplement(method string) error {
	return &types.AdapterError{Code: types.ErrSystemException, Msg: method + " not implement"}
}
