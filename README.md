# wallet-adapter

**github.com/godaddy-x/wallet-adapter** — 多主链钱包适配器基础库，提供交易单构建、广播与区块扫描等能力。同系列子类库如 **github.com/godaddy-x/wallet-adapter-eth** 等可依赖本库实现各链适配。

## 能力概览

- **统一交易类型**：`RawTransaction`、`PendingSignTx`（待签名交易单）、`Transaction`、`SummaryRawTransaction`
- **核心流程**：入口为 flow（`adapter.BuildTransaction` 创建待签名交易单 PendingSignTx、`adapter.SendTransaction` 验证+广播）；decoder 只负责构建/验签/提交 rawTx，签名由外部 MPC 完成
- **区块扫描**：`BlockScanner` 接口与 `BlockScannerBase`，支持按高度扫描区块、持续扫块循环、补扫单高度、提取交易与回执
- **链抽象**：`ChainAdapter`、`TransactionDecoder`、`BlockScanner`、`AddressDecoder`；可选 `WalletDAI` 回调查询钱包/账户/地址等
- **链配置**：`config` 包提供 `Configer` 接口与 INI 解析（`KVFromINIFile`/`KVFromINIContent`、`MapConfig`），供 `AssetsConfig.LoadAssetsConfig` 等复用
- **智能合约**（可选）：`SmartContractDecoder` 与 `ABIDAI`：代币余额、ABI 调用/创建/广播、合约元数据；`GetSmartContractDecoder(symbol)` 按链获取
- **多链注册**：按 symbol 注册/查询：`RegAdapter`、`GetAdapter`、`GetTransactionDecoder`、`GetBlockScanner`、`GetAddressDecoder`、`GetSmartContractDecoder`

## 项目结构（按 package 划分）

```
wallet-adapter/
├── go.mod
├── README.md
├── doc.go                    # 根包 adapter 说明
├── export.go                 # 类型与函数统一导出
├── types/                    # 数据类型与错误码
│   ├── types.go              # 交易、账户、地址等核心类型
│   ├── errors.go             # 错误码与 AdapterError
│   ├── symbol.go             # 链/币种信息 SymbolInfo
│   ├── contract.go           # 智能合约相关 TokenBalance、SmartContractRawTransaction、SmartContractCallResult、ABIInfo
│   └── block.go              # 扫块相关 BlockHeader、TxExtractData、ExtractDataItem、ContractReceiptItem、Balance、UnscanRecord、SmartContractReceipt 等
├── wallet/                   # 钱包数据访问接口（与 types 解耦）
│   └── wallet.go             # Wallet、WalletDAI、WalletDAIBase
├── decoder/                  # 解码器（交易 + 地址 + 智能合约）
│   ├── transaction.go        # TransactionDecoder、TransactionDecoderBase
│   ├── address.go            # AddressDecoder、AddressDecoderBase
│   └── contract.go           # SmartContractDecoder、SmartContractDecoderBase、ABIDAI
├── config/                   # 链配置通用接口与 INI 解析
│   ├── configer.go           # Configer、MapConfig（供 LoadAssetsConfig 使用）
│   └── ini.go                # KVFromINIFile、KVFromINIContent
├── chain/                    # 链适配器与注册表
│   ├── adapter.go            # ChainAdapter、ChainAdapterBase
│   ├── config.go             # AssetsConfig、AssetsConfigBase
│   └── registry.go           # RegAdapter、GetAdapter、GetTransactionDecoder 等
├── flow/                     # 构建与广播流程（入口：BuildTransaction/BuildSummaryTransaction/SendTransaction）
│   └── flow.go               # 调 decoder 构建 rawTx，再调 wrapper.SignPendingTxData 得 PendingSignTx；广播前校验 DataSign/TradeSign
├── scanner/                  # 区块扫描器
│   ├── scanner.go            # BlockScanner 接口与 Base（按高度扫块、持续循环、插队扫描、地址余额查询）
│   └── SCANNER.md            # 扫块器详细设计文档
```

- **推荐**：`import "github.com/godaddy-x/wallet-adapter"` 后使用 `adapter.BuildTransaction`、`adapter.RawTransaction`、`adapter.RegAdapter` 等。
- **按需引用子包**：如 `import "github.com/godaddy-x/wallet-adapter/types"`、`import "github.com/godaddy-x/wallet-adapter/wallet"`、`import "github.com/godaddy-x/wallet-adapter/decoder"`、`import "github.com/godaddy-x/wallet-adapter/config"`、`import "github.com/godaddy-x/wallet-adapter/chain"` 等。

## 接入新链步骤

1. **实现 `TransactionDecoder`**
   - `CreateRawTransaction`：根据 `RawTransaction` 构建 rawHex/fees/sigParts 等（只负责构建 rawTx）
   - `CreateSummaryRawTransactionWithError`：汇总场景生成多笔 `RawTransactionWithError`
   - `VerifyRawTransaction`：校验 rawTx（合并 SignerList 后）签名
   - `SubmitRawTransaction`：提交已签名 rawTx 到链上节点并返回 `Transaction`
   - 可选：`SignRawTransaction`（本地签名时实现）、`GetRawTransactionFeeRate`、`EstimateRawTransactionFee`

2. **实现 `ChainAdapter`**
   - 实现 `SymbolInfo`（Symbol、Decimal 等）
   - `GetTransactionDecoder()` 返回上述 decoder
   - （可选）`GetAddressDecoder()` 返回地址解析器；`GetBlockScanner()` 返回扫块器；`GetSmartContractDecoder()` 返回智能合约解析器

3. **注册**
   - 在 `init()` 或启动时：`adapter.RegAdapter("SYMBOL", yourAdapter)`

4. **（可选）实现 `BlockScanner`**
   - 嵌入 `scanner.Base`，实现 `ScanBlockWithResult`（按高度扫块并返回结果）、`ScanBlockOnce`（单高度补扫）、`RunScanLoop`（持续扫块循环）、`ScanBlockPrioritize`（插队扫描）、`ResetScanHeight`（重置游标）、`GetBalanceByAddress`（地址余额查询）等。
   - 通过 `SetBlockScanTargetFunc` 设置扫描目标查询，`SetTokenMetadataFunc` 注入合约元数据查询，供扫块时补充合约信息。
   - `GetBalanceByAddress` 可使用 `QueryBalancesConcurrent` 辅助函数实现并发查询。

5. **（可选）实现 `AddressDecoder`**
   - 嵌入 `decoder.AddressDecoderBase`，按需实现：`PublicKeyToAddress`、`AddressVerify`、`AddressDecode`、`AddressEncode`、WIF、多签、`CustomCreateAddress` 等；未实现的方法由 Base 返回“未实现”。

6. **（可选）实现 `SmartContractDecoder`**（见 `decoder/contract.go`）
   - 嵌入 `decoder.SmartContractDecoderBase`，按需实现：`GetTokenBalanceByAddress`、`CallSmartContractABI`、`CreateSmartContractRawTransaction`、`SubmitSmartContractRawTransaction`、`GetABIInfo`、`SetABIInfo`、`GetTokenMetadata`；链不支持合约则 `GetSmartContractDecoder()` 返回 nil。

## 使用示例

```go
import "github.com/godaddy-x/wallet-adapter"

// 1. 获取某链的 TransactionDecoder（需已 RegAdapter）
decoder, err := adapter.GetTransactionDecoder("BTC")
if err != nil {
    return err
}

// 2. 构造原始交易单
rawTx := &adapter.RawTransaction{
    Coin:    adapter.Coin{Symbol: "BTC"},
    Account: account,           // *adapter.AssetsAccount
    To:      map[string]string{toAddress: amount},
    FeeRate: feeRate,
    Required: 1,
}

// 3. 调用 flow 构建待签名交易单（decoder 构建 rawTx → wrapper.SignPendingTxData 填 DataSign/TradeSign → 返回 PendingSignTx）
//    wrapper 实现 adapter.WalletDAI，BuildTransaction/SendTransaction 时不可为 nil
pendingTx, err := adapter.BuildTransaction(decoder, wrapper, rawTx)
if err != nil {
    return err
}
// ... 调用 MPC 签名，填充 pendingTx.SignerList ...

// 4. 广播（内部会复算 DataSign/TradeSign 校验 Data 未被篡改，再验签并提交）
tx, err := adapter.SendTransaction(decoder, wrapper, pendingTx)
if err != nil {
    return err
}
// tx.TxID, tx.Status 等
```

本库为 **github.com/godaddy-x** 下的基础适配器模块，以**币类转账**为主（交易构建/广播、区块扫描、地址解析），并可选扩展**智能合约**（`decoder/contract.go` 的 `SmartContractDecoder`、`types/contract.go` 的合约相关类型）。链实现（如 github.com/godaddy-x/wallet-adapter-eth）、MPC 签名库等可依赖本库；不包含 HD 钱包、具体链实现等。

## License

见项目根目录 LICENSE。
