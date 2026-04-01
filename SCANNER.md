### `scanner` 扫块器设计文档

`scanner` 包提供一套可复用的**区块扫描框架**，用于从多条链上扫描新区块、提取交易/合约回执，并通过同步回调返回扫描结果。

---

## 1. 角色与职责

- **BlockScanner 接口**  
  抽象"一条链的扫块能力"，包括：
  - 按高度扫块（`ScanBlockWithResult`、`ScanBlockOnce`）
  - 持续扫块循环（`RunScanLoop`，参数使用 `ScanLoopParams` 结构体，便于扩展）
  - 插队扫描（`ScanBlockPrioritize`）
  - 游标重置（`ResetScanHeight`）
  - 交易 / 合约回执提取
  - 地址余额查询（`GetBalanceByAddress`）
  - 入账前复核（`VerifyTransactionByTxID`、`VerifyTransactionMatch`）
  - 注入扫描目标函数

- **Base 基类** (`scanner.Base`)  
  提供**基础注入与默认未实现方法**：
  - `ScanTargetFunc` 注入（扫描目标查询）
  - 所有 `BlockScanner` 接口方法的默认"未实现"返回

  每条链的扫块实现嵌入 `*scanner.Base`，按需重写接口方法。

- **types/block.go 中的扫块相关类型**
  - `BlockHeader`：区块头
  - `TxExtractData`：交易提取结果
  - `ExtractDataItem`：按 SourceKey 聚合的交易提取结果项
  - `ContractReceiptItem`：合约回执项
  - `SmartContractReceipt` / `SmartContractEvent`：合约回执与事件
  - `ScanTargetParam`：扫描目标查询
  - `BlockScanResult` / `TxVerifyResult` / `TxVerifyMatchResult`：扫块与复核结果

---

## 2. BlockScanner 核心接口

```go
type BlockScanner interface {
    // 扫描目标：根据地址/别名/公钥等筛选业务关心的交易
    SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

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
    // 参数使用 ScanLoopParams 结构体，后续添加新字段无需修改方法签名
    RunScanLoop(params ScanLoopParams) error

    // ScanBlockPrioritize 插队扫描指定高度列表（用于补扫/漏扫紧急修复）。
    // 说明：插队高度在 RunScanLoop 主线扫描间隙优先处理，结果通过 params.HandleBlock 回调。
    ScanBlockPrioritize(heights []uint64) error
}

// ScanLoopParams RunScanLoop 的参数结构体
// 后续添加新参数无需修改方法签名，直接在此结构体添加字段即可
type ScanLoopParams struct {
    StartHeight    uint64                           // 起始扫描高度，从 StartHeight+1 开始扫描
    Confirmations  uint64                           // 确认数，仅用于计算 BlockHeader.Confirmations 供业务层参考
    Interval       time.Duration                    // 每轮扫描后的休眠间隔
    HandleBlock    func(res *types.BlockScanResult) // 每扫完一个高度的回调函数（可为 nil）
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
// BlockScanTargetFunc 批量查询扫描目标所属信息，供扫块时过滤交易。
// 调用方会传入单个批次参数（同 Symbol + ScanTargetType），回调在对象上原地填充结果：
//   - 命中目标：ScanTarget[target] 写入非 nil 值（地址建议写 accountID string，合约建议写 *types.Coin）
//   - 未命中目标：保持 ScanTarget[target]=nil
type BlockScanTargetFunc func(target *types.ScanTargetParam) error
```

- `ScanTargetParam` 描述"想要关注什么"：
  - `Symbol`：链标识
  - `ScanTarget`：目标集合（`map[string]interface{}`），key 为地址/别名/合约地址/公钥/备注，value 为命中结果（`nil` 未命中，非 `nil` 命中）
  - `ScanTargetType`：类型枚举

- 命中值约定：
  - 地址目标：写入 `accountID string`
  - 合约目标：写入 `*types.Coin`（`IsContract=true` 且包含完整 `Contract` 元数据）

在扫描每笔交易时，链实现可通过注入的 `ScanTargetFunc` 进行过滤。

### 3.2 合约元数据返回方式

- 合约场景建议在 `ScanTargetFunc` 中，当 `ScanTargetType == ScanTargetTypeContractAddress` 时，
  直接写入 `*types.Coin`（或 `types.Coin`）。
- 链实现会直接读取 `Coin.Contract` 填充交易中的合约信息。

---

## 4. Base 基类行为

`scanner.Base` 主要职责：

- 提供 `ScanTargetFunc` 的注入方法
- 为所有 `BlockScanner` 接口方法提供默认"未实现"返回
- 提供 `Run` / `Pause` / `Stop` / `Restart` 运行控制方法
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
   - 重写 `ScanBlockPrioritize`：插队扫描指定高度（可选，默认返回未实现错误）
   - 重写 `VerifyTransactionByTxID` / `VerifyTransactionMatch`：入账前复核
   - 重写 `GetCurrentBlockHeader` / `GetGlobalMaxBlockHeight`：状态查询
   - 重写 `GetBalanceByAddress`：查询地址余额（使用 `QueryBalancesConcurrent` 辅助函数）

2. **注入依赖**
   - 调用 `SetBlockScanTargetFunc` 注入批量扫描目标查询函数（推荐在一次回调内完成批量 DB 查询并原地回填）

3. **在链适配器中挂载**
   - 在 `chain.Adapter` 实现中，`GetBlockScanner()` 返回该链的 `BlockScanner` 实例
