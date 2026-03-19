### `scanner` 扫块器设计文档

`scanner` 包提供一套可复用的**区块扫描框架**，用于从多条链上扫描新区块、提取交易/合约回执，并通过同步回调返回扫描结果。

---

## 1. 角色与职责

- **BlockScanner 接口**  
  抽象"一条链的扫块能力"，包括：
  - 按高度扫块（`ScanBlockWithResult`、`ScanBlockOnce`）
  - 持续扫块循环（`RunScanLoop`）
  - 游标重置（`ResetScanHeight`）
  - 交易 / 合约回执提取
  - 入账前复核（`VerifyTransactionByTxID`、`VerifyTransactionMatch`）
  - 注入扫描目标函数和合约元数据查询

- **Base 基类** (`scanner.Base`)  
  提供**基础注入与默认未实现方法**：
  - `ScanTargetFunc` 注入（扫描目标查询）
  - `TokenMetadataFunc` 注入（合约元数据查询）
  - 所有 `BlockScanner` 接口方法的默认"未实现"返回

  每条链的扫块实现嵌入 `*scanner.Base`，按需重写接口方法。

- **types/block.go 中的扫块相关类型**
  - `BlockHeader`：区块头
  - `TxExtractData`：交易提取结果
  - `ExtractDataItem`：按 SourceKey 聚合的交易提取结果项
  - `ContractReceiptItem`：合约回执项
  - `SmartContractReceipt` / `SmartContractEvent`：合约回执与事件
  - `ScanTargetParam` / `ScanTargetResult`：扫描目标查询
  - `BlockScanResult` / `TxVerifyResult` / `TxVerifyMatchResult`：扫块与复核结果

---

## 2. BlockScanner 核心接口

```go
type BlockScanner interface {
    // 扫描目标：根据地址/别名/公钥等筛选业务关心的交易
    SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

    // 合约元数据：用于补全 token decimals 等信息（建议外部注入，避免扫块器自行猜测默认值）
    SetTokenMetadataFunc(tokenMetadataFunc TokenMetadataFunc) error

    // 运行控制：启动/停止内部扫描任务
    Run() error
    Pause() error

    // ScanBlockWithResult 按高度扫描区块并返回摘要结果，供外部系统推进游标与重试
    ScanBlockWithResult(height uint64) (*types.BlockScanResult, error)
    // ScanBlockOnce 指定高度扫描一次（用于补扫/漏扫修复），不进入持续循环
    ScanBlockOnce(height uint64) (*types.BlockScanResult, error)

    // ResetScanHeight 将持续扫块循环的起始高度重置到指定值（用于回滚重扫）
    ResetScanHeight(height uint64) error

    // 状态查询
    GetCurrentBlockHeader() (*types.BlockHeader, error)
    GetGlobalMaxBlockHeight() uint64

    // 交易 / 回执提取
    ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) ([]*types.ExtractDataItem, []*types.ContractReceiptItem, error)

    // GetBalanceByAddress 查询指定地址的余额
    GetBalanceByAddress(address ...string) ([]*types.Balance, error)

    // VerifyTransactionByTxID 入账前按 txid 二次复核链上结果并返回可入账结果集
    VerifyTransactionByTxID(txid string, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyResult, error)

    // VerifyTransactionMatch 入账前对链上结果集做二次复核，并与外部期望对象严格比对
    VerifyTransactionMatch(txid string, expected *types.TxVerifyExpected, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyMatchResult, error)

    // RunScanLoop 按高度持续扫描区块，回调每个高度的扫描结果给外部系统
    RunScanLoop(startHeight, confirmations, windowSize uint64, interval time.Duration, handleBlock func(res *types.BlockScanResult)) error
}
```

每条链实现自己的 `BlockScanner`，建议模式：

```go
type MyChainScanner struct {
    *scanner.Base
    // ... 链特有字段（RPC 客户端等）
}

func NewMyChainScanner() *MyChainScanner {
    return &MyChainScanner{Base: scanner.NewBlockScannerBase()}
}

// 按需重写 BlockScanner 接口方法
func (s *MyChainScanner) ScanBlockWithResult(height uint64) (*types.BlockScanResult, error) {
    // 1. 从节点获取区块
    // 2. 解析交易/回执
    // 3. 过滤业务关心的目标
    // 4. 返回 BlockScanResult
}
```

---

## 3. 扫描目标与合约元数据

### 3.1 扫描目标函数

```go
type BlockScanTargetFunc func(target types.ScanTargetParam) types.ScanTargetResult
```

- `ScanTargetParam` 描述"想要关注什么"：
  - `Symbol`：链标识
  - `ScanTarget`：地址 / 别名 / 合约地址 / 公钥 / 备注
  - `ScanTargetType`：类型枚举

- `ScanTargetResult` 返回：
  - `SourceKey`：业务侧自定义源标识（例如钱包 ID）
  - `Exist`：该目标是否存在 / 被订阅

在扫描每笔交易时，链实现可通过注入的 `ScanTargetFunc` 进行过滤。

### 3.2 TokenMetadataFunc：合约元数据查询

```go
type TokenMetadataFunc func(symbol, contractAddr string) *types.SmartContract
```

- 由外部注入，用于根据链标识和合约地址查询代币/合约元数据
- 扫块器在解析合约回执时，通过 `TokenMetadataFunc` 填充 `Contract` 字段
- 注入方式示例：

```go
bs := scanner.NewBlockScannerBase()
bs.SetTokenMetadataFunc(func(symbol, contractAddr string) *types.SmartContract {
    // 从缓存 / 配置 / 远程服务查询
    return lookupTokenMetadata(symbol, contractAddr)
})
```

`TokenMetadataFunc` 是**可选依赖**：未设置时，链实现应在使用前做 nil 判断。

---

## 4. Base 基类行为

`scanner.Base` 主要职责：

- 提供 `ScanTargetFunc` 和 `TokenMetadataFunc` 的注入方法
- 为所有 `BlockScanner` 接口方法提供默认"未实现"返回
- 提供 `QueryBalancesConcurrent` 辅助函数，用于并发查询地址余额

链实现嵌入 `*scanner.Base` 后，按需重写需要的方法即可。

### 4.1 使用 QueryBalancesConcurrent 实现 GetBalanceByAddress

```go
// GetBalanceByAddress 并发查询多个地址余额
func (bs *MyChainScanner) GetBalanceByAddress(address ...string) ([]*types.Balance, error) {
    // 定义查询函数：查询单个地址的已确认、未确认、总余额
    queryFunc := func(addr string) (confirmed, unconfirmed, total string, err error) {
        // 调用链 RPC 查询余额
        balanceConfirmed, err := bs.GetAddrBalanceFromNode(addr, "latest")
        if err != nil {
            return "", "", "", err
        }
        
        // pending 状态包含未确认的交易
        balanceAll, err := bs.GetAddrBalanceFromNode(addr, "pending")
        if err != nil {
            balanceAll = balanceConfirmed
        }
        
        // 计算未确认余额
        unconfirmedBI := new(big.Int).Sub(balanceAll, balanceConfirmed)
        
        return 
            ConvertToDecimal(balanceConfirmed),  // confirmed
            ConvertToDecimal(unconfirmedBI),     // unconfirmed
            ConvertToDecimal(balanceAll),        // total
            nil
    }
    
    // 使用 Base 提供的并发查询辅助函数
    return bs.QueryBalancesConcurrent(bs.Symbol(), address, queryFunc, 20)
}
```

---

## 5. 接入新链的扫块实现步骤

1. **实现 `BlockScanner`**
   - 定义结构体，嵌入 `*scanner.Base`
   - 重写 `ScanBlockWithResult`：按高度扫描区块、解析交易/回执、过滤目标、返回结果
   - 重写 `ScanBlockOnce`：单高度补扫逻辑（可复用 `ScanBlockWithResult`）
   - 重写 `RunScanLoop`：持续扫块循环（或在外部实现循环逻辑）
   - 重写 `VerifyTransactionByTxID` / `VerifyTransactionMatch`：入账前复核
   - 重写 `GetCurrentBlockHeader` / `GetGlobalMaxBlockHeight`：状态查询
   - 重写 `GetBalanceByAddress`：查询地址余额（可使用 `QueryBalancesConcurrent` 辅助函数）

2. **注入依赖**
   - 调用 `SetBlockScanTargetFunc` 注入扫描目标查询函数
   - 调用 `SetTokenMetadataFunc` 注入合约元数据查询函数（可选）

3. **在链适配器中挂载**
   - 在 `chain.Adapter` 实现中，`GetBlockScanner()` 返回该链的 `BlockScanner` 实例
