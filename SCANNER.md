### `scanner` 扫块器设计文档

`scanner` 包提供一套可复用的**区块扫描框架**，用于从多条链上扫描新区块、提取交易/合约回执，并通过观察者模式通知业务侧。

---

## 1. 角色与职责

- **BlockScanner 接口**  
  抽象“一条链的扫块能力”，包括：
  - 定时扫块任务控制（`Run`/`Stop`/`Pause`/`Restart`）
  - 按高度扫块与查询当前区块头
  - 交易 / 合约回执提取
  - 地址余额与地址交易查询
  - 区块链同步状态查询
  - 注入扫描目标函数和观察者

- **Base 基类** (`scanner.Base`)  
  提供**通用控制逻辑**：
  - 定时任务调度（`taskRunner` + `PeriodOfTask`）
  - 观察者注册/移除
  - 通过带缓冲 channel 转发新区块通知
  - 生命周期管理（`InitBlockScanner` / `CloseBlockScanner`）
  - 注入 `BlockchainDAI`

  每条链的扫块实现通常嵌入 `*scanner.Base`，只重写与链相关的方法。

- **BlockchainDAI 接口**  
  抽象**区块链数据持久化访问层**，与具体 DB/缓存解耦：
  - 当前扫描高度（区块头）的持久化与读取
  - 本地区块头缓存
  - 扫描失败记录（未扫记录）的存取
  - 交易结果落库
  - 区块头缓存上限配置

  配套基类 `BlockchainDAIBase` 所有方法默认返回“未实现”，实现方按需重写。

- **types/block.go 中的类型**  
  定义与扫块相关的统一数据结构，例如：
  - `BlockHeader`、`UnscanRecord`、`Balance`
  - `Recharge` / `TxInput` / `TxOutPut`
  - `TxExtractData`、`ScanTargetParam` / `ScanTargetResult`
  - `BlockchainSyncStatus`
  - `SmartContractReceipt` / `SmartContractEvent`

---

## 2. BlockScanner 核心接口（已精简）

```go
type BlockScanner interface {
    // 扫描目标：根据地址/别名/公钥等筛选业务关心的交易
    SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

    // 合约元数据：用于补全 token decimals 等信息（建议外部注入，避免扫块器自行猜测默认值）
    SetTokenMetadataFunc(tokenMetadataFunc TokenMetadataFunc) error

    // 扫块（兼容旧异步通知模式：实现方可内部向观察者推送）
    ScanBlock(height uint64) error
    // ScanBlockWithResult 按高度扫描区块并返回摘要结果，供外部系统推进游标与重试。
    ScanBlockWithResult(height uint64) (*types.BlockScanResult, error)

    // 状态查询
    GetCurrentBlockHeader() (*types.BlockHeader, error)
    GetGlobalMaxBlockHeight() uint64

    // 交易 / 回执提取
    ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*types.TxExtractData, map[string]*types.SmartContractReceipt, error)

    // VerifyTransactionByTxID 入账前按 txid 二次复核链上结果并返回可入账结果集。
    VerifyTransactionByTxID(txid string, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyResult, error)

    // VerifyTransactionMatch 入账前对链上结果集做二次复核，并与外部期望对象 expected 严格比对。
    VerifyTransactionMatch(txid string, expected *types.TxVerifyExpected, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyMatchResult, error)

    // RunScanLoop 外部维护游标的持续扫块循环（回调每个高度的扫描结果）。
    RunScanLoop(startHeight, confirmations, windowSize uint64, interval time.Duration, handleBlock func(res *types.BlockScanResult)) error
}
```

每条链实现自己的 `BlockScanner`，建议模式：

```go
type MyChainScanner struct {
    *scanner.Base
    // ... 链特有字段
}

func NewMyChainScanner() *MyChainScanner {
    bs := scanner.NewBlockScannerBase()
    s  := &MyChainScanner{Base: bs}

    // 设置定时任务：例如每 PeriodOfTask 扫一次块
    s.SetTask(func() {
        // 1. 查询当前高度 / 网络最新高度
        // 2. 逐块调用 s.ScanBlock(height)
        // 3. 持久化进度并调用 s.NewBlockNotify(header)
    })
    return s
}
```

链实现需重写：`ScanBlock`、`GetCurrentBlockHeader`、`GetGlobalMaxBlockHeight`、`Extract*`、`GetBalanceByAddress` 等。

---

## 3. 扫描目标与观察者

### 3.1 扫描目标函数

```go
type BlockScanTargetFunc func(target types.ScanTargetParam) types.ScanTargetResult
```

- `ScanTargetParam` 描述“想要关注什么”：
  - `Symbol`：链标识
  - `ScanTarget`：地址 / 别名 / 合约地址 / 公钥 / 备注
  - `ScanTargetType`：类型枚举（`ScanTargetTypeAccountAddress`、`ScanTargetTypeContractAddress` 等）

- `ScanTargetResult` 返回：
  - `SourceKey`：业务侧自定义源标识（例如钱包 ID、账户 ID）
  - `Exist`：该目标是否存在 / 被订阅
  - `TargetInfo`：可选附加信息（业务自定义结构）

在扫描每笔交易时，链实现可通过注入的 `ScanTargetFunc` 进行过滤，只对业务关心的目标提取并推送结果。

### 3.2 TokenMetadataFunc：合约元数据查询

```go
type TokenMetadataFunc func(symbol, contractAddr string) *types.SmartContract
```

- 由外部（通常是链适配器或业务层）注入，用于**根据链标识和合约地址查询代币/合约元数据**。
- 扫块器在解析交易或合约回执时，可以通过 `TokenMetadataFunc` 填充 `SmartContractReceipt`、`SmartContractEvent` 中的 `Contract` 等字段，避免在扫块器内部耦合具体的元数据来源（本地缓存、远程服务等）。
- 注入方式示例：

```go
bs := scanner.NewBlockScannerBase()
bs.SetTokenMetadataFunc(func(symbol, contractAddr string) *types.SmartContract {
    // 从缓存 / 配置 / 远程服务中查询合约元数据
    return lookupTokenMetadata(symbol, contractAddr)
})
```

`TokenMetadataFunc` 是**可选依赖**：未设置时，链实现应在使用前做 nil 判断，或保持合约元数据为空。

### 3.3 观察者接口

```go
type BlockScanNotificationObject interface {
    BlockScanNotify(header *types.BlockHeader) error
    BlockExtractDataNotify(sourceKey string, data *types.TxExtractData) error
    BlockExtractSmartContractDataNotify(sourceKey string, data *types.SmartContractReceipt) error
}
```

- `BlockScanNotify`：收到新区块头（例如用于 UI 或内部监控）
- `BlockExtractDataNotify`：通知某源下的新交易提取结果（含 TxInputs/TXOutputs/Transaction）
- `BlockExtractSmartContractDataNotify`：通知某源下的新合约回执

`scanner.Base` 内部使用两个带缓冲的 channel：

- `blockProducer`：`NewBlockNotify` 写入区块头
- `blockConsumer`：独立 goroutine 消费并 fan-out 至所有观察者

**注意**：`NewBlockNotify` 是**非阻塞**的——当 `blockProducer` 满时，会直接丢弃后续通知，以避免慢观察者阻塞扫块协程。这意味着：

- 扫块和数据持久化不受影响；
- 但**实时通知可能丢失部分块事件**，业务如需“绝不丢通知”，应在上层追加可靠队列或适当增大缓冲并接受阻塞。

---

## 4. BlockchainDAI：区块链数据访问接口

```go
type BlockchainDAI interface {
    // 扫描进度
    SaveCurrentBlockHead(header *types.BlockHeader) error
    GetCurrentBlockHead(symbol string) (*types.BlockHeader, error)

    // 本地区块头缓存
    SaveLocalBlockHead(header *types.BlockHeader) error
    GetLocalBlockHeadByHeight(height uint64, symbol string) (*types.BlockHeader, error)

    // 未扫记录
    SaveUnscanRecord(record *types.UnscanRecord) error
    DeleteUnscanRecordByHeight(height uint64, symbol string) error
    DeleteUnscanRecordByID(id, symbol string) error
    GetUnscanRecords(symbol string) ([]*types.UnscanRecord, error)

    // 缓存控制 & 交易落库
    SetMaxBlockCache(max uint64, symbol string) error
    SaveTransaction(tx *types.Transaction) error
}
```

用途：

- 将**当前已扫描到的最新区块头**持久化，重启后可从断点恢复。
- 将**未扫成功的区块或交易**记录下来，以便重试或人工处理。
- 将**扫描出的交易结果**落库，供业务查询。

`BlockchainDAIBase` 给出了所有方法的默认“未实现”实现，链适配器可按需选择性实现。

---

## 5. Base 基类行为

`scanner.Base` 主要职责：

- 管理观察者集合（线程安全）
- 管理定时任务（`taskRunner`）
- 管理扫块状态（`isClose`、`scanning`）
- 通过 channel 承载新区块通知并 fan-out

关键点：

- `NewBlockScannerBase()` 会：
  - 初始化 `Observers` map 与 `PeriodOfTask`（默认 5 秒）
  - 调用 `InitBlockScanner()` 启动通知转发协程

- `SetTask(task func())`：
  - 设置或替换当前定时任务（内部使用 `time.Ticker`）
  - `Run()` 将启动 `taskRunner`，定期执行该任务

- `CloseBlockScanner()`：
  - 先标记 `isClose = true`，再停止任务，再安全关闭 `blockProducer` channel，避免 `NewBlockNotify` 向已关闭通道写入。

链实现只要专注于“如何按高度从链上取块、解析交易/回执并写入 DAI”，其他控制逻辑交给 `Base` 即可。

---

## 6. 接入新链的扫块实现步骤（简版）

1. **实现 `BlockchainDAI`（可选但强烈建议）**
   - 负责持久化区块头、未扫记录、交易结果。
   - 嵌入 `BlockchainDAIBase`，只重写实际需要的方法。

2. **实现链专用 `BlockScanner`**
   - 定义结构体，嵌入 `*scanner.Base`：

```go
type MyChainScanner struct {
    *scanner.Base
    // RPC 客户端、配置等
}
```

   - 在构造函数中：
     - 调用 `scanner.NewBlockScannerBase()`
     - 注入 `BlockchainDAI`（如有）
     - 设置 `SetTask` 指向“从当前高度向上扫描”的逻辑

3. **实现核心方法**
   - `ScanBlock(height uint64) error`：按高度从节点拉取区块、解析交易/合约回执、调用 `ScanTargetFunc` 过滤、写入 DAI，并在成功后调用 `NewBlockNotify(header)`。
   - `GetCurrentBlockHeader()` / `GetGlobalMaxBlockHeight()` / `GetScannedBlockHeight()`：用于状态查询和同步控制。
   - `ExtractTransactionData` / `ExtractTransactionAndReceiptData`：按交易 ID 复查或补扫时使用。
   - `GetBalanceByAddress` / `GetTransactionsByAddress`：为上层提供地址维度的余额与交易查询。

4. **在链适配器中挂载扫描器**
   - 在 `chain.Adapter` 的具体实现中，`GetBlockScanner()` 返回该链的 `BlockScanner` 实例。

