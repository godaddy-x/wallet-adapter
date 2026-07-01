# wallet-adapter

**github.com/godaddy-x/wallet-adapter** — Multi-chain wallet adapter foundation library providing transaction building, broadcast, and block scanning. Chain-specific libraries such as **github.com/godaddy-x/wallet-adapter-eth** depend on this module to implement per-chain adapters.

## Capabilities

- **Unified transaction types**: `RawTransaction`, `PendingSignTx` (pending sign payload), `Transaction`, `SummaryRawTransaction`
- **Batch transfer (legacy type=0)**: `BatchRawRequest` remains in `types/`; **production batch** now uses scanner `CreateBatchTransferTrade` → `SmartContractRawTransaction` (CLI type=2). `adapter.BuildBatchTransaction` is no longer the production path
- **Core flow**: Entry points in `flow` (`adapter.BuildTransaction` creates `PendingSignTx`, `adapter.SendTransaction` verifies and broadcasts); decoders only build/verify/submit rawTx — signing is done externally (e.g. MPC)
- **Block scanning**: `BlockScanner` interface and `BlockScannerBase` — scan by height, continuous loop, single-height catch-up, extract transactions and receipts
- **Chain abstraction**: `ChainAdapter`, `TransactionDecoder`, `BlockScanner`, `AddressDecoder`; optional `WalletDAI` callback for wallet/account/address/balance lookups
- **Chain config**: `config` package provides `Configer` and JSON parsing (`KVFromJSONFile` / `KVFromJSONContent`, `MapConfig`) for `AssetsConfig.LoadAssetsConfig` and similar
- **Smart contracts** (optional): `SmartContractDecoder` and `ABIDAI` — token balance, ABI call/create/broadcast, contract metadata; `GetSmartContractDecoder(symbol)` per chain
- **Multi-chain registry**: Register and resolve by symbol — `RegAdapter`, `GetAdapter`, `GetTransactionDecoder`, `GetBlockScanner`, `GetAddressDecoder`, `GetSmartContractDecoder`

## Project layout (by package)

```
wallet-adapter/
├── go.mod
├── README.md
├── SCANNER.md                # Block scanner design doc
├── doc.go                    # Root adapter package docs
├── export.go                 # Unified type and function exports
├── types/                    # Data types and error codes
│   ├── types.go              # Core types: transactions, accounts, addresses
│   ├── errors.go             # Error codes and AdapterError
│   ├── symbol.go             # Chain/asset SymbolInfo
│   ├── batch_raw.go          # Batch transfer BatchRawRequest, BatchTransferRecipient (v1.0.4+)
│   ├── contract.go           # Smart contract: TokenBalance, SmartContractRawTransaction, etc.
│   └── block.go              # Scanning: BlockHeader, TxExtractData, ExtractDataItem, etc.
├── wallet/                   # Wallet data access (decoupled from types)
│   └── wallet.go             # WalletDAI, WalletDAIBase (callbacks for flow/decoder)
├── decoder/                  # Decoders (transaction + address + smart contract)
│   ├── transaction.go        # TransactionDecoder, TransactionDecoderBase
│   ├── address.go            # AddressDecoder, AddressDecoderBase
│   └── contract.go           # SmartContractDecoder, SmartContractDecoderBase, ABIDAI
├── config/                   # Chain config interface and JSON parsing
│   ├── configer.go           # Configer, MapConfig (for LoadAssetsConfig)
│   └── json.go               # KVFromJSONFile, KVFromJSONContent
├── chain/                    # Chain adapter and registry
│   ├── adapter.go            # ChainAdapter, ChainAdapterBase
│   ├── config.go             # AssetsConfig, AssetsConfigBase
│   └── registry.go           # RegAdapter, GetAdapter, GetTransactionDecoder, etc.
├── flow/                     # Build and broadcast (BuildTransaction / SendTransaction)
│   └── flow.go               # BuildTransaction / SendTransaction; DataSign/TradeSign checks before broadcast
└── scanner/                  # Block scanner
    └── scanner.go            # BlockScanner interface and Base
```

- **Recommended**: `import "github.com/godaddy-x/wallet-adapter"` then use `adapter.BuildTransaction`, `adapter.RawTransaction`, `adapter.RegAdapter`, etc.
- **Subpackages as needed**: e.g. `import "github.com/godaddy-x/wallet-adapter/types"`, `wallet`, `decoder`, `config`, `chain`, etc.

## Adding a new chain

1. **Implement `TransactionDecoder`**
   - `CreateRawTransaction`: build rawHex/fees/sigParts from `RawTransaction`
   - `CreateSummaryRawTransactionWithError`: summary flow producing multiple `RawTransactionWithError`
   - `VerifyRawTransaction`: verify signed rawTx (after merging `SignerList`)
   - `SubmitRawTransaction`: submit signed rawTx to the node and return `Transaction`
   - Optional: `SignRawTransaction` (local signing), `GetRawTransactionFeeRate`, `EstimateRawTransactionFee`

2. **Implement `ChainAdapter`**
   - Implement `SymbolInfo` (Symbol, Decimal, etc.)
   - `GetTransactionDecoder()` returns the decoder above
   - Optional: `GetAddressDecoder()`, `GetBlockScanner()`, `GetSmartContractDecoder()`

3. **Register**
   - In `init()` or at startup: `adapter.RegAdapter("SYMBOL", yourAdapter)`

4. **(Optional) Implement `BlockScanner`**
   - Embed `scanner.Base`; implement `ScanBlockWithResult`, `ScanBlockOnce`, `RunScanLoop`, `ScanBlockPrioritize`, `ResetScanHeight`, `GetBalanceByAddress`, etc.
   - Use `SetBlockScanTargetFunc` for batch scan target lookup (`*ScanTargetParam` with `ScanTarget` as `map[string]interface{}`: `nil` = miss, non-`nil` = hit; address hits should use `accountID string`, contract hits `*types.Coin`).
   - `GetBalanceByAddress` can use helper `QueryBalancesConcurrent`.

5. **(Optional) Implement `AddressDecoder`**
   - Embed `decoder.AddressDecoderBase`; implement as needed: `PublicKeyToAddress`, `AddressVerify`, `AddressDecode`, `AddressEncode`, WIF, multisig, `CustomCreateAddress`, etc. Unimplemented methods return “not implemented” from Base.

6. **(Optional) Implement `SmartContractDecoder`** (see `decoder/contract.go`)
   - Embed `decoder.SmartContractDecoderBase`; implement as needed: `GetTokenBalanceByAddress`, `CallSmartContractABI`, `CreateSmartContractRawTransaction`, `SubmitSmartContractRawTransaction`, `GetABIInfo`, `SetABIInfo`, `GetTokenMetadata`. Return nil from `GetSmartContractDecoder()` if the chain has no contract support.

See [SCANNER.md](SCANNER.md) for block scanner design details.

## Usage example

```go
import "github.com/godaddy-x/wallet-adapter"

// 1. Get TransactionDecoder for a chain (must have called RegAdapter)
decoder, err := adapter.GetTransactionDecoder("BTC")
if err != nil {
    return err
}

// 2. Build raw transaction
rawTx := &adapter.RawTransaction{
    Coin:    adapter.Coin{Symbol: "BTC"},
    Account: account,           // *adapter.AssetsAccount
    To:      map[string]string{toAddress: amount},
    FeeRate: feeRate,
    Required: 1,
}

// 3. Build pending sign tx via flow (decoder builds rawTx → wrapper.SignPendingTxData fills DataSign/TradeSign)
//    wrapper must implement adapter.WalletDAI; cannot be nil for BuildTransaction/SendTransaction
pendingTx, err := adapter.BuildTransaction(decoder, wrapper, rawTx)
if err != nil {
    return err
}
// ... MPC signing, fill pendingTx.SignerList ...

// 4. Broadcast (recomputes DataSign/TradeSign, verifies Data was not tampered, then verify sign and submit)
tx, err := adapter.SendTransaction(decoder, wrapper, pendingTx)
if err != nil {
    return err
}
// tx.TxID, tx.Status, etc.
```

This is the **github.com/godaddy-x** foundation adapter module focused on **asset transfers** (tx build/broadcast, block scan, address decode) with optional **smart contract** extensions (`decoder/contract.go`, `types/contract.go`). Chain implementations (e.g. wallet-adapter-eth) and MPC signing libraries depend on it; HD wallets and concrete chain logic are out of scope.

## License

See LICENSE in the repository root.
