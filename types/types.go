// Package types 多主链适配器核心数据类型（交易、账户、错误等）
package types

// Coin 币种/链标识
//
//easyjson:json
type Coin struct {
	Symbol     string        `json:"symbol"`
	IsContract bool          `json:"isContract"`
	ContractID string        `json:"contractID"`
	Contract   SmartContract `json:"contract"`
}

// SmartContract 智能合约/代币信息
//
//easyjson:json
type SmartContract struct {
	ContractID string `json:"contractID"`
	Symbol     string `json:"symbol"`
	Address    string `json:"address"`
	Token      string `json:"token"`
	Protocol   string `json:"protocol"`
	Name       string `json:"name"`
	Decimals   uint64 `json:"decimals"`
}

// PendingSignTx 待签名交易单：构建后产出，广播时入参。
// 构建成功即确定 Data（原始交易单 JSON，贯穿流程不变，仅反序列化读参）、DataSign、TradeSign（用于保证 Data 不被篡改）；
// 随后由 MPC 签名填充 SignerList，再提交广播。
//
//easyjson:json
type PendingSignTx struct {
	Sid        string            `json:"sid"`
	Data       string            `json:"data"`       // 原始交易单 JSON，全程不变
	DataSign   string            `json:"dataSign"`   // 构建时确定，保证 Data 不被篡改
	TradeSign  string            `json:"tradeSign"`  // 构建时确定，保证 Data 不被篡改
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	SignerList map[string]string `json:"signerList"` // MPC 签名后填充，再提交广播
}

// RawTransaction 原始交易单
//
//easyjson:json
type RawTransaction struct {
	Coin        Coin                       `json:"coin"`
	TxID        string                     `json:"txID"`
	RawHex      string                     `json:"rawHex"`
	FeeRate     string                     `json:"feeRate"`
	To          map[string]string          `json:"to"`
	Account     *AssetsAccount             `json:"account"`
	Signatures  map[string][]*KeySignature `json:"sigParts"`
	Required    uint64                     `json:"reqSigs"`
	IsBuilt     bool                       `json:"isBuilt"`
	IsCompleted bool                       `json:"isComplete"`
	IsSubmit    bool                       `json:"isSubmit"`
	Change      *Address                   `json:"change"`
	ExtParam    string                     `json:"extParam"`

	Sid         string   `json:"sid"`
	CreateTime  int64    `json:"createTime"`
	CreateNonce string   `json:"createNonce"`
	TxType      int64    `json:"txType"`
	Fees        string   `json:"fees"`
	TxAmount    string   `json:"txAmount"`
	TxFrom      []string `json:"txFrom"`
	TxTo        []string `json:"txTo"`
}

// KeySignature 单笔签名
//
//easyjson:json
type KeySignature struct {
	EccType   uint32   `json:"eccType"`
	Nonce     string   `json:"nonce"`
	Address   *Address `json:"address"`
	Signature string   `json:"signed"`
	Message   string   `json:"msg"`
	RSV       bool     `json:"rsv"`
}

// Transaction 广播后的交易结果
//
//easyjson:json
type Transaction struct {
	ID          string   `json:"id"`
	WxID        string   `json:"wxid"`
	TxID        string   `json:"txid"`
	AccountID   string   `json:"accountID"`
	Coin        Coin     `json:"coin"`
	From        []string `json:"from"`
	To          []string `json:"to"`
	Amount      string   `json:"amount"`
	Decimal     int32    `json:"decimal"`
	TxType      uint64   `json:"txType"`
	TxAction    string   `json:"txAction"`
	Confirm     int64    `json:"confirm"`
	BlockHash   string   `json:"blockHash"`
	BlockHeight uint64   `json:"blockHeight"`
	Fees        string   `json:"fees"`
	SubmitTime  int64    `json:"submitTime"`
	ConfirmTime int64    `json:"confirmTime"`
	Status      string   `json:"status"`
	Reason      string   `json:"reason"`
	ExtParam    string   `json:"extParam"`
}

const (
	TxStatusSuccess = "1"
	TxStatusFail    = "0"
)

// SummaryRawTransaction 汇总交易参数
//
//easyjson:json
type SummaryRawTransaction struct {
	Sid                string              `json:"sid"`
	Coin               Coin                `json:"coin"`
	FeeRate            string              `json:"feeRate"`
	SummaryAddress     string              `json:"summaryAddress"`
	MinTransfer        string              `json:"minTransfer"`
	RetainedBalance    string              `json:"retainedBalance"`
	Account            *AssetsAccount      `json:"account"`
	AddressStartIndex  int                 `json:"addressStartIndex"`
	AddressLimit       int                 `json:"addressLimit"`
	Confirms           uint64              `json:"confirms"`
	FeesSupportAccount *FeesSupportAccount `json:"feesSupportAccount"`
	ExtParam           string              `json:"extParam"`
}

// FeesSupportAccount 手续费支持账户
//
//easyjson:json
type FeesSupportAccount struct {
	AccountID        string `json:"accountID"`
	FixSupportAmount string `json:"fixSupportAmount"`
	FeesSupportScale string `json:"feesScale"`
}

// RawTransactionWithError 带错误的原始交易（汇总场景）
//
//easyjson:json
type RawTransactionWithError struct {
	RawTx *RawTransaction `json:"rawTx"`
	Error *AdapterError   `json:"error"`
}

// AssetsAccount 资产账户
//
//easyjson:json
type AssetsAccount struct {
	ID        string   `json:"id"`
	WalletID  string   `json:"walletID"`
	Alias     string   `json:"alias"`
	AccountID string   `json:"accountID"`
	Index     uint64   `json:"index"`
	HDPath    string   `json:"hdPath"`
	PublicKey string   `json:"publicKey"`
	OwnerKeys []string `json:"ownerKeys"`
	Required  uint64   `json:"required"`
	Symbol    string   `json:"symbol"`
	Balance   string   `json:"balance"`
	ExtParam  string   `json:"extParam"`
}

// Address 地址
//
//easyjson:json
type Address struct {
	ID        string `json:"id"`
	AccountID string `json:"accountID"`
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Symbol    string `json:"symbol"`
	Balance   string `json:"balance"`
	Index     uint64 `json:"index"`
	HDPath    string `json:"hdPath"`
	ExtParam  string `json:"extParam"`
}
