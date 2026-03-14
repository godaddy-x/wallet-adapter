// Package adapter 多主链适配器基础框架，提供交易构建、广播与区块扫描等核心能力。
//
// 子包划分：
//   - types   — 数据类型与错误码（含 BlockHeader、TxExtractData、Balance 等扫块类型）
//   - wallet  — 钱包数据访问：Wallet、WalletDAI、WalletDAIBase（供 flow/decoder 回调查询钱包/账户/地址等）
//   - decoder — 解码器：TransactionDecoder（交易单）、AddressDecoder（地址），各带 Base 基类；签名由外部 MPC 提供
//   - chain   — 链适配器 ChainAdapter 与注册表 RegAdapter/GetAdapter/GetTransactionDecoder/GetBlockScanner/GetAddressDecoder
//   - flow    — 构建与广播流程 BuildTransaction、BuildSummaryTransaction、SendTransaction（可传入 WalletDAI 回调查询）
//   - scanner — 区块扫描器 BlockScanner、BlockchainDAI 与 Base
//
// 本包对上述子包做统一导出，便于调用方 import "github.com/blockchain/wallet-adapter" 使用。
//
// 使用示例：
//
//	decoder, _ := adapter.GetTransactionDecoder("BTC")
//	rawTx := &adapter.RawTransaction{ Coin: adapter.Coin{Symbol: "BTC"}, Account: account, To: map[string]string{toAddr: amount}, Required: 1 }
//	pendingTx, _ := adapter.BuildTransaction(decoder, wrapper, rawTx) // 入口为 flow，返回 PendingSignTx；wrapper 不可为 nil
//	tx, _ := adapter.SendTransaction(decoder, wrapper, pendingTx)
package adapter
