// Package wallet 钱包数据访问接口（WalletDAI），供 TransactionDecoder 在构建/验证时回调查询，与 types 解耦。
package wallet

import "github.com/godaddy-x/wallet-adapter/types"

type SearchParams struct {
	CountQ          bool   // 是否查询总条数（为 true 时返回的 int64 为总条数）
	WalletID        string // 钱包 ID，用于过滤
	AccountID       string // 账户 ID，用于过滤
	Address         string // 地址，用于过滤
	LastID          int64  // 上一页最后一条 ID，0 表示首页
	Limit           int64  // 每页条数
	ContractAddress string // 合约地址 用于过滤
	MinTransfer     string // 筛选余额最小值
	LastTransfer    string // 最后记录余额值
}

// WalletDAI 钱包数据访问接口，供 flow 与 TransactionDecoder 在构建/验证时回调查询。
// 调用 flow.BuildTransaction / BuildSummaryTransaction / SendTransaction 时 wrapper 不可为 nil（需调用 SignPendingTxData 或校验）；
// decoder 内部回调查询时 wrapper 可为 nil 表示不回查。实现方仅实现需要的方法即可，未实现由 WalletDAIBase 返回“未实现”。
type WalletDAI interface {
	// GetAssetsAccountInfo 根据账户 ID 查询资产账户信息。params.AccountID 为要查询的账户 ID。
	GetAssetsAccountInfo(params SearchParams) (*types.AssetsAccount, error)
	// GetAssetsAccountList 分页查询资产账户列表。params 包含分页和查询条件：CountQ 是否查询总条数（为 true 时返回的 int64 为总条数）；WalletID/AccountID/Address 过滤条件；LastID 为上一页最后一条 ID，0 表示首页；Limit 每页条数。
	GetAssetsAccountList(params SearchParams) ([]*types.AssetsAccount, int64, error)
	// GetAssetsAccountByAddress 根据地址查询所属资产账户。params.Address 为要查询的地址。
	GetAssetsAccountByAddress(params SearchParams) (*types.AssetsAccount, error)
	// GetAddress 根据地址查询地址详情（含公钥、标签等）。params.Address 为要查询的地址。
	GetAddress(params SearchParams) (*types.Address, error)
	// GetAddressList 分页查询地址列表。params 包含分页和查询条件：CountQ 是否查询总条数（为 true 时返回的 int64 为总条数）；WalletID/AccountID/Address 过滤条件；LastID 为上一页最后一条 ID，0 表示首页；Limit 每页条数。
	GetAddressList(params SearchParams) ([]*types.Address, int64, error)
	// GetAccountBalanceList 查询帐户余额
	GetAccountBalanceList(params SearchParams) ([]*types.AssetBalance, int64, error)
	// GetAddressBalanceList 查询地址余额
	GetAddressBalanceList(params SearchParams) ([]*types.AssetBalance, int64, error)
	// SetAddressExtParam 设置地址扩展参数，key 为业务自定义键。
	SetAddressExtParam(address string, key string, val interface{}) error
	// GetAddressExtParam 获取地址扩展参数。
	GetAddressExtParam(address string, key string) (interface{}, error)
	// GetTransactionByTxID 根据交易 ID 与链标识查询交易记录，用于校验、补扫等。
	// 返回的 Transaction 使用分离字段设计：
	//   - FromAddr/FromAmt: 发送方地址和金额列表（一一对应）
	//   - ToAddr/ToAmt: 接收方地址和金额列表（一一对应）
	//   - TxAction: 交易方向标记（"send"=转出, "receive"=转入, "internal"=内部转账, "fee"=手续费）
	//   - OutputIndex: 输出索引（-2=手续费, -1=主币, 0+=合约事件索引）
	GetTransactionByTxID(txID, symbol string) ([]*types.Transaction, error)
	// SignPendingTxData 对原始交易 JSON（rawTx）做业务侧签名，返回填充了 DataSign/TradeSign 的 PendingSignTx，用于保证创建交易单数据有效性；
	// SendTransaction 广播前会再次调用本方法复算 DataSign/TradeSign，用于校验 Data 未被篡改。
	SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error)
}

// WalletDAIBase 为 WalletDAI 的默认未实现基类，所有方法均返回“未实现”错误；
// 实现方嵌入此结构体并仅重写需要的方法即可。
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
func (WalletDAIBase) SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error) {
	return nil, errNotImplement("SignPendingTxData")
}

func errNotImplement(method string) error {
	return &types.AdapterError{Code: types.ErrSystemException, Msg: method + " not implement"}
}
