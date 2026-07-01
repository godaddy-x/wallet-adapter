# Block scanner design (`scanner` package)

The `scanner` package provides a reusable **block scanning framework** for scanning new blocks on multiple chains, extracting transactions/contract receipts, and returning results via synchronous callbacks.

---

## 1. Roles and responsibilities

- **`BlockScanner` interface**  
  Abstracts per-chain scanning:
  - Scan by height (`ScanBlockWithResult`, `ScanBlockOnce`)
  - Continuous scan loop (`RunScanLoop` with `ScanLoopParams` for extensibility)
  - Priority catch-up (`ScanBlockPrioritize`)
  - Cursor reset (`ResetScanHeight`)
  - Transaction / contract receipt extraction
  - Address balance lookup (`GetBalanceByAddress`)
  - Pre-credit verification (`VerifyTransactionByTxID`, `VerifyTransactionMatch`)
  - Scan target function injection

- **`Base` (`scanner.Base`)**  
  Provides **injection and default “not implemented” stubs**:
  - `ScanTargetFunc` injection
  - Default “not implemented” for all `BlockScanner` methods

  Each chain embeds `*scanner.Base` and overrides methods as needed.

- **Scan-related types in `types/block.go`**
  - `BlockHeader`
  - `TxExtractData`
  - `ExtractDataItem` (aggregated by `SourceKey`)
  - `ContractReceiptItem`
  - `SmartContractReceipt` / `SmartContractEvent`
  - `ScanTargetParam`
  - `BlockScanResult` / `TxVerifyResult` / `TxVerifyMatchResult`

---

## 2. Core `BlockScanner` interface

```go
type BlockScanner interface {
    // Scan targets: filter business-relevant txs by address/alias/pubkey etc.
    SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

    // Run control: start/stop internal scan task
    Run() error
    Pause() error

    // ScanBlockWithResult scans one height and returns summary for cursor advance / retry
    ScanBlockWithResult(height uint64) (*types.BlockScanResult, error)
    // ScanBlockOnce scans one height once (catch-up / gap repair), not part of continuous loop
    ScanBlockOnce(height uint64) (*types.BlockScanResult, error)

    // ResetScanHeight resets continuous loop start height (re-scan from rollback point)
    ResetScanHeight(height uint64) error

    // Status
    GetCurrentBlockHeader() (*types.BlockHeader, error)
    GetGlobalMaxBlockHeight() uint64

    // Transaction / receipt extraction
    ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) ([]*types.ExtractDataItem, []*types.ContractReceiptItem, error)

    // GetBalanceByAddress queries balances for given addresses
    GetBalanceByAddress(address ...string) ([]*types.Balance, error)

    // VerifyTransactionByTxID re-verifies on-chain result by txid before crediting
    VerifyTransactionByTxID(txid string, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyResult, error)

    // VerifyTransactionMatch re-verifies on-chain set and strictly compares with external expectation
    VerifyTransactionMatch(txid string, expected *types.TxVerifyExpected, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyMatchResult, error)

    // RunScanLoop continuously scans by height; callbacks per height via params.HandleBlock
    // ScanLoopParams allows new fields without changing the method signature
    RunScanLoop(params ScanLoopParams) error

    // ScanBlockPrioritize priority-scans given heights (urgent catch-up).
    // Priority heights are handled between main-loop scans in RunScanLoop; results via params.HandleBlock.
    ScanBlockPrioritize(heights []uint64) error
}

// ScanLoopParams parameters for RunScanLoop
type ScanLoopParams struct {
    StartHeight    uint64                           // loop starts from StartHeight+1
    Confirmations  uint64                           // used only for BlockHeader.Confirmations hint
    Interval       time.Duration                    // sleep between rounds
    HandleBlock    func(res *types.BlockScanResult) // per-height callback (may be nil)
}
```

Recommended pattern per chain:

```go
type MyChainScanner struct {
    *scanner.Base
    // ... chain-specific fields (RPC client, etc.)
}

func NewMyChainScanner() *MyChainScanner {
    return &MyChainScanner{Base: scanner.NewBlockScannerBase()}
}

func (s *MyChainScanner) ScanBlockWithResult(height uint64) (*types.BlockScanResult, error) {
    // 1. Fetch block from node
    // 2. Parse txs/receipts
    // 3. Filter business targets
    // 4. Return BlockScanResult
}
```

---

## 3. Scan targets and contract metadata

### 3.1 Scan target function

```go
// BlockScanTargetFunc batch-queries scan targets for filtering during scan.
// Caller passes one batch (same Symbol + ScanTargetType); callback fills results in place:
//   - hit: ScanTarget[target] = non-nil (address → accountID string; contract → *types.Coin)
//   - miss: ScanTarget[target] remains nil
type BlockScanTargetFunc func(target *types.ScanTargetParam) error
```

- `ScanTargetParam` describes what to watch:
  - `Symbol`: chain id
  - `ScanTarget`: `map[string]interface{}` — keys are address/alias/contract/pubkey/memo; values are hit results (`nil` = miss)
  - `ScanTargetType`: type enum

- Hit value conventions:
  - Address target: `accountID string`
  - Contract target: `*types.Coin` (`IsContract=true` with full `Contract` metadata)

Chain implementations filter each tx using the injected `ScanTargetFunc`.

### 3.2 Contract metadata

- For contracts, when `ScanTargetType == ScanTargetTypeContractAddress`, write `*types.Coin` (or `types.Coin`) in `ScanTargetFunc`.
- The chain implementation reads `Coin.Contract` into the transaction.
- **ERC20 token contracts** (`ScanTargetTypeContractAddress`): lookup **token contract config** on the business side; hit value is `*types.Coin` (including `decimals`).
- **BatchSender batch contracts** (`ScanTargetTypeBatchSenderAddress`): lookup **confirmed deployment binding** (`category=batch_transfer`), separate from token config; hit value is non-nil (e.g. `true` or deployment record pointer).
- Do **not** mix the two contract kinds in one table or `ScanTargetType`; chain adapters branch on `ScanTargetType`, not `Protocol`.

---

## 4. `Base` behavior

`scanner.Base` mainly:

- Injects `ScanTargetFunc`
- Default “not implemented” for all `BlockScanner` methods
- Provides `Run` / `Pause` / `Stop` / `Restart`
- Provides `QueryBalancesConcurrent` for concurrent balance queries

Embed `*scanner.Base` and override only what you need.

### 4.1 Implementing `GetBalanceByAddress` with `QueryBalancesConcurrent`

```go
func (bs *MyChainScanner) GetBalanceByAddress(address ...string) ([]*types.Balance, error) {
    queryFunc := func(addr string) (confirmed, unconfirmed, total string, err error) {
        balanceConfirmed, err := bs.GetAddrBalanceFromNode(addr, "latest")
        if err != nil {
            return "", "", "", err
        }

        balanceAll, err := bs.GetAddrBalanceFromNode(addr, "pending")
        if err != nil {
            balanceAll = balanceConfirmed
        }

        unconfirmedBI := new(big.Int).Sub(balanceAll, balanceConfirmed)

        return
            ConvertToDecimal(balanceConfirmed),
            ConvertToDecimal(unconfirmedBI),
            ConvertToDecimal(balanceAll),
            nil
    }

    return bs.QueryBalancesConcurrent(bs.Symbol(), address, queryFunc, 20)
}
```

---

## 5. Adding block scanning for a new chain

1. **Implement `BlockScanner`**
   - Struct embedding `*scanner.Base`
   - Override `ScanBlockWithResult`: fetch block, parse txs/receipts, filter targets, return result
   - Override `ScanBlockOnce`: single-height catch-up (can delegate to `ScanBlockWithResult`)
   - Override `RunScanLoop`: continuous loop (or implement loop externally)
   - Override `ScanBlockPrioritize`: optional priority heights (default returns not implemented)
   - Override `VerifyTransactionByTxID` / `VerifyTransactionMatch`: pre-credit verification
   - Override `GetCurrentBlockHeader` / `GetGlobalMaxBlockHeight`: status
   - Override `GetBalanceByAddress`: use `QueryBalancesConcurrent` if helpful

2. **Inject dependencies**
   - Call `SetBlockScanTargetFunc` with batch DB lookup that fills `ScanTarget` in place

3. **Wire into chain adapter**
   - `GetBlockScanner()` on `ChainAdapter` returns the chain’s `BlockScanner` instance
