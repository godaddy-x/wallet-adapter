// Package adapter multi-chain adapter foundation providing core capabilities such as transaction building, broadcasting, and block scanning.
//
// Subpackages:
//   - types   — data types and error codes (including block-scan types like BlockHeader, TxExtractData, Balance)
//   - wallet  — wallet data access: Wallet, WalletDAI, WalletDAIBase (for flow/decoder callbacks to query wallet/account/address/balance, etc.)
//   - decoder — decoders: TransactionDecoder (transaction.go), AddressDecoder (address.go), SmartContractDecoder (contract.go, optional), each with a Base class; signing is provided by external MPC
//   - config  — common chain config interface Configer, MapConfig, and JSON parsing (KVFromJSONFile, KVFromJSONContent), reused by AssetsConfig.LoadAssetsConfig
//   - chain   — chain adapter ChainAdapter and registry RegAdapter/GetAdapter/GetTransactionDecoder/GetBlockScanner/GetAddressDecoder/GetSmartContractDecoder; AssetsConfig and SmartContractDecoder are optional
//   - flow    — build and broadcast flow: BuildTransaction, BuildSmartContractTransaction, BuildSummaryTransaction, SendTransaction (WalletDAI may be passed for callback queries)
//   - scanner — block scanner BlockScanner and Base (scan by height, continuous loop, priority scan, address balance queries)
//   - amount  — on-chain amount precision conversion (human-readable ↔ smallest on-chain unit); downstream import "github.com/godaddy-x/wallet-adapter/amount"
//
// This package re-exports the subpackages above so callers can import "github.com/godaddy-x/wallet-adapter".
//
// Example:
//
//	decoder, _ := adapter.GetTransactionDecoder("BTC")
//	rawTx := &adapter.RawTransaction{ Coin: adapter.Coin{Symbol: "BTC"}, Account: account, To: map[string]string{toAddr: amount}, Required: 1 }
//	pendingTx, _ := adapter.BuildTransaction(decoder, wrapper, rawTx) // entry is flow; returns PendingSignTx; wrapper must not be nil
//	tx, _ := adapter.SendTransaction(decoder, wrapper, pendingTx)
package adapter
