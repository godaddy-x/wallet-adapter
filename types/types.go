// Package types core multi-chain adapter data types (transactions, accounts, errors, etc.)
package types

// Coin coin/chain identifier
//
//easyjson:json
type Coin struct {
	Symbol     string        `json:"symbol"`
	IsContract bool          `json:"isContract"`
	Contract   SmartContract `json:"contract"`
}

// SmartContract smart contract/token info
//
//easyjson:json
type SmartContract struct {
	Symbol   string `json:"symbol"`
	Address  string `json:"address"`
	Token    string `json:"token"`
	Protocol string `json:"protocol"` // contract role; see SmartContractProtocol* constants
	Name     string `json:"name"`
	Decimals uint64 `json:"decimals"`
}

// ScanTarget contract hit value (*types.Coin / types.Coin) SmartContract.Protocol convention:
// written by business layer when registering contracts; used by chain adapters to distinguish ERC20 token contracts from BatchSender batch transfers, etc.
const (
	SmartContractProtocolERC20       = "erc20"
	SmartContractProtocolBatchSender = "batch_sender"
)

// PendingSignTx pending-sign transaction: produced after build, input when broadcasting.
// On successful build, Data (raw transaction JSON, immutable throughout the flow, only deserialized for reads), DataSign, and TradeSign are fixed (to prevent Data tampering);
// MPC then fills SignerList before submit/broadcast.
//
//easyjson:json
type PendingSignTx struct {
	Sid        string            `json:"sid"`
	Data       string            `json:"data"`      // raw transaction JSON, immutable throughout
	DataSign   string            `json:"dataSign"`  // fixed at build time to prevent Data tampering
	TradeSign  string            `json:"tradeSign"` // fixed at build time to prevent Data tampering
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	SignerList map[string]string `json:"signerList"` // filled by MPC before submit/broadcast
}

// RawTransaction raw transaction
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
	ExtParam    map[string]string          `json:"extParam"`
	// SignExt signature verification metadata (JSON object string); written only by adapter at build time; includes chainId, signScheme, etc.
	SignExt string `json:"signExt,omitempty"`
	// SpeedUp speed up/replace pending transaction; when non-empty, build uses fixed nonce and explicit gas, not automatic nonce/gas.
	SpeedUp *SpeedUp `json:"speedUp,omitempty"`

	Sid         string   `json:"sid"`
	CreateTime  int64    `json:"createTime"`
	CreateNonce string   `json:"createNonce"`
	TxType      int64    `json:"txType"`
	Fees        string   `json:"fees"`
	TxAmount    string   `json:"txAmount"`
	TxFrom      []string `json:"txFrom"`
	TxTo        []string `json:"txTo"`
}

// KeySignature single signature
//
//easyjson:json
type KeySignature struct {
	EccType   uint32   `json:"eccType"` // 1.ecdsa 2.ed25519
	Nonce     string   `json:"nonce"`
	Address   *Address `json:"address"`
	Signature string   `json:"signed"`
	Message   string   `json:"msg"`
	RSV       bool     `json:"rsv"`
}

// Transaction broadcast transaction result
//
//easyjson:json
type Transaction struct {
	ID        string `json:"id"`
	WxID      string `json:"wxid"`
	TxID      string `json:"txid"`
	AccountID string `json:"accountID"`
	Coin      Coin   `json:"coin"`
	// FromAddr sender address list, one-to-one with FromAmt
	FromAddr []string `json:"fromAddr"`
	// FromAmt sender amount list, one-to-one with FromAddr
	FromAmt []string `json:"fromAmt"`
	// ToAddr receiver address list, one-to-one with ToAmt
	ToAddr []string `json:"toAddr"`
	// ToAmt receiver amount list, one-to-one with ToAddr
	ToAmt       []string `json:"toAmt"`
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
	// OutputIndex output index for this Transaction; semantics vary by chain:
	// - EVM chains: event log index (logIndex); >=0 contract event, -1 native transfer, -2 fee record
	// - UTXO chains (BTC, etc.): transaction output index (vout)
	// Used to pinpoint records within the same tx for business-layer uniqueness.
	OutputIndex int64 `json:"outputIndex"`

	// FeeType marks fee-class records (e.g. "gas").
	// Empty means a normal transfer record.
	FeeType string `json:"feeType"`
	// ExtParam key-value extension fields for structured extra info (e.g. contract_creation).
	ExtParam map[string]string `json:"extParam"`
}

const (
	TxStatusSuccess = "1"
	TxStatusFail    = "0"
)

// SummaryRawTransaction summary transaction parameters
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
	AddressStartIndex  int64               `json:"addressStartIndex"`
	AddressLimit       int64               `json:"addressLimit"`
	Confirms           uint64              `json:"confirms"`
	FeesSupportAccount *FeesSupportAccount `json:"feesSupportAccount"`
	ExtParam           string              `json:"extParam"`
}

// FeesSupportAccount fee support account
//
//easyjson:json
type FeesSupportAccount struct {
	AccountID        string `json:"accountID"`
	FixSupportAmount string `json:"fixSupportAmount"`
	FeesSupportScale string `json:"feesScale"`
}

// RawTransactionWithError raw transaction with error (summary scenario)
//
//easyjson:json
type RawTransactionWithError struct {
	RawTx *RawTransaction `json:"rawTx"`
	Error *AdapterError   `json:"error"`
}

// AssetsAccount asset account
//
//easyjson:json
type AssetsAccount struct {
	ID        int64    `json:"id"`
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

// Address address
//
//easyjson:json
type Address struct {
	ID        int64  `json:"id"`
	AccountID string `json:"accountID"`
	Address   string `json:"address"`
	PublicKey string `json:"publicKey"`
	Symbol    string `json:"symbol"`
	Balance   string `json:"balance"`
	Index     uint64 `json:"index"`
	HDPath    string `json:"hdPath"`
	ExtParam  string `json:"extParam"`
}

// AssetBalance account/address asset balance
//
//easyjson:json
type AssetBalance struct {
	ID               int64  `json:"id"`
	WalletID         string `json:"walletID"`
	AccountID        string `json:"accountID"`
	Address          string `json:"address"`
	MainSymbol       string `json:"mainSymbol"` // base category -> mainSymbol
	Symbol           string `json:"symbol"`
	ContractAddress  string `json:"contractAddress"`
	Balance          string `json:"balance"`
	ConfirmBalance   string `json:"confirmBalance"`
	UnconfirmBalance string `json:"unconfirmBalance"`
}
