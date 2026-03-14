// Package scanner 区块扫描器接口与基类，负责扫描新区块并向观察者推送订阅地址的新交易
package scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/blockchain/wallet-adapter/types"
)

// BlockScanTargetFunc 根据扫描目标参数（地址/别名等）查询所属源与是否存在，供扫块时过滤交易。
type BlockScanTargetFunc func(target types.ScanTargetParam) types.ScanTargetResult

// BlockScanNotificationObject 扫描通知观察者：新区块、交易提取结果、合约回执。
type BlockScanNotificationObject interface {
	BlockScanNotify(header *types.BlockHeader) error
	BlockExtractDataNotify(sourceKey string, data *types.TxExtractData) error
	BlockExtractSmartContractDataNotify(sourceKey string, data *types.SmartContractReceipt) error
}

// BlockScanner 区块扫描器接口：设置扫描目标、观察者、扫块任务，执行 ScanBlock/GetCurrentBlockHeader，提取交易与余额。
type BlockScanner interface {
	SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

	AddObserver(obj BlockScanNotificationObject) error
	RemoveObserver(obj BlockScanNotificationObject) error

	SetRescanBlockHeight(height uint64) error
	Run() error
	Stop() error
	Pause() error
	Restart() error

	InitBlockScanner() error
	CloseBlockScanner() error

	ScanBlock(height uint64) error
	NewBlockNotify(header *types.BlockHeader) error

	GetCurrentBlockHeader() (*types.BlockHeader, error)
	GetGlobalMaxBlockHeight() uint64
	GetScannedBlockHeight() uint64

	ExtractTransactionData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*types.TxExtractData, error)
	ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*types.TxExtractData, map[string]*types.SmartContractReceipt, error)

	GetBalanceByAddress(address ...string) ([]*types.Balance, error)
	GetTransactionsByAddress(offset, limit int, coin types.Coin, address ...string) ([]*types.TxExtractData, error)

	SetBlockchainDAI(dai BlockchainDAI) error
	SupportBlockchainDAI() bool

	GetBlockchainSyncStatus() (*types.BlockchainSyncStatus, error)
}

// BlockchainDAI 区块链数据访问接口：区块头、未扫记录、按 txID 查询交易及持久化，供扫描器与业务解耦存储。
// 实现方可按需实现部分方法，未实现由 BlockchainDAIBase 返回“未实现”。
type BlockchainDAI interface {
	// SaveCurrentBlockHead 持久化当前已扫到的链上最新区块头（扫描进度）。
	SaveCurrentBlockHead(header *types.BlockHeader) error
	// GetCurrentBlockHead 获取指定链已持久化的当前区块头。
	GetCurrentBlockHead(symbol string) (*types.BlockHeader, error)
	// SaveLocalBlockHead 持久化本地缓存的区块头（如按高度缓存的块头）。
	SaveLocalBlockHead(header *types.BlockHeader) error
	// GetLocalBlockHeadByHeight 根据高度与链标识获取本地缓存的区块头。
	GetLocalBlockHeadByHeight(height uint64, symbol string) (*types.BlockHeader, error)
	// SaveUnscanRecord 保存未扫记录（扫块失败或需重试时写入）。
	SaveUnscanRecord(record *types.UnscanRecord) error
	// DeleteUnscanRecordByHeight 按高度与链标识删除未扫记录（重扫成功后清理）。
	DeleteUnscanRecordByHeight(height uint64, symbol string) error
	// DeleteUnscanRecordByID 按记录 ID 与链标识删除未扫记录。
	DeleteUnscanRecordByID(id, symbol string) error
	// GetTransactionsByTxID 根据交易 ID 与链标识查询已持久化的交易，用于去重、补扫等。
	GetTransactionsByTxID(txID, symbol string) ([]*types.Transaction, error)
	// GetUnscanRecords 获取指定链的未扫记录列表，用于重试或人工处理。
	GetUnscanRecords(symbol string) ([]*types.UnscanRecord, error)
	// SetMaxBlockCache 设置指定链的本地区块头缓存最大数量（如按高度缓存的块数上限）。
	SetMaxBlockCache(max uint64, symbol string) error
	// SaveTransaction 持久化单笔交易（扫块解析出的交易落库）。
	SaveTransaction(tx *types.Transaction) error
}

// BlockchainDAIBase 为 BlockchainDAI 的默认未实现基类，所有方法均返回“未实现”错误；
// 实现方嵌入此结构体并仅重写需要的方法即可。
type BlockchainDAIBase struct{}

func (BlockchainDAIBase) SaveCurrentBlockHead(*types.BlockHeader) error {
	return fmt.Errorf("SaveCurrentBlockHead not implement")
}
func (BlockchainDAIBase) GetCurrentBlockHead(string) (*types.BlockHeader, error) {
	return nil, fmt.Errorf("GetCurrentBlockHead not implement")
}
func (BlockchainDAIBase) SaveLocalBlockHead(*types.BlockHeader) error {
	return fmt.Errorf("SaveLocalBlockHead not implement")
}
func (BlockchainDAIBase) GetLocalBlockHeadByHeight(uint64, string) (*types.BlockHeader, error) {
	return nil, fmt.Errorf("GetLocalBlockHeadByHeight not implement")
}
func (BlockchainDAIBase) SaveUnscanRecord(*types.UnscanRecord) error {
	return fmt.Errorf("SaveUnscanRecord not implement")
}
func (BlockchainDAIBase) DeleteUnscanRecordByHeight(uint64, string) error {
	return fmt.Errorf("DeleteUnscanRecordByHeight not implement")
}
func (BlockchainDAIBase) DeleteUnscanRecordByID(string, string) error {
	return fmt.Errorf("DeleteUnscanRecordByID not implement")
}
func (BlockchainDAIBase) GetTransactionsByTxID(string, string) ([]*types.Transaction, error) {
	return nil, fmt.Errorf("GetTransactionsByTxID not implement")
}
func (BlockchainDAIBase) GetUnscanRecords(string) ([]*types.UnscanRecord, error) {
	return nil, fmt.Errorf("GetUnscanRecords not implement")
}
func (BlockchainDAIBase) SetMaxBlockCache(uint64, string) error {
	return fmt.Errorf("SetMaxBlockCache not implement")
}
func (BlockchainDAIBase) SaveTransaction(*types.Transaction) error {
	return fmt.Errorf("SaveTransaction not implement")
}

const (
	defaultPeriodOfTask   = 5 * time.Second
	blockNotifyBufferSize = 64 // 新区块通知管道缓冲，避免观察者稍慢时扫块循环被阻塞
)

// Base 区块扫描器基类：观察者、ScanTargetFunc、定时任务、BlockchainDAI、NewBlockNotify 转发；Run/Stop/Pause/Restart；具体链需实现 ScanBlock、GetCurrentBlockHeader、Extract*、GetBalanceByAddress 等。
type Base struct {
	Mu             sync.RWMutex
	Observers      map[BlockScanNotificationObject]bool
	ScanTargetFunc BlockScanTargetFunc
	PeriodOfTask   time.Duration
	BlockchainDAI  BlockchainDAI

	blockProducer chan interface{}
	blockConsumer chan interface{}
	isClose       bool
	scanning      bool
	taskRunner    *taskRunner
}

type taskRunner struct {
	f    func()
	tick time.Duration
	stop chan struct{}
}

func (t *taskRunner) Start() {
	if t == nil || t.f == nil {
		return
	}
	t.stop = make(chan struct{})
	ticker := time.NewTicker(t.tick)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-t.stop:
				return
			case <-ticker.C:
				t.f()
			}
		}
	}()
}

func (t *taskRunner) Stop() {
	if t != nil && t.stop != nil {
		close(t.stop)
		t.stop = nil
	}
}

func (t *taskRunner) Running() bool { return t != nil && t.stop != nil }

// NewBlockScannerBase 创建扫描器基类
func NewBlockScannerBase() *Base {
	bs := &Base{
		Observers:    make(map[BlockScanNotificationObject]bool),
		PeriodOfTask: defaultPeriodOfTask,
	}
	bs.InitBlockScanner()
	return bs
}

func (bs *Base) InitBlockScanner() error {
	bs.blockProducer = make(chan interface{}, blockNotifyBufferSize)
	bs.blockConsumer = make(chan interface{}, blockNotifyBufferSize)
	bs.isClose = false
	go bs.forwardBlockNotify()
	go bs.consumeBlockNotify()
	return nil
}

func (bs *Base) forwardBlockNotify() {
	for v := range bs.blockProducer {
		bs.blockConsumer <- v
	}
	close(bs.blockConsumer)
}

func (bs *Base) consumeBlockNotify() {
	for obj := range bs.blockConsumer {
		header, ok := obj.(*types.BlockHeader)
		if !ok {
			continue
		}
		bs.Mu.RLock()
		observers := make([]BlockScanNotificationObject, 0, len(bs.Observers))
		for o := range bs.Observers {
			observers = append(observers, o)
		}
		bs.Mu.RUnlock()
		for _, o := range observers {
			_ = o.BlockScanNotify(header)
		}
	}
}

func (bs *Base) SetBlockScanTargetFunc(f BlockScanTargetFunc) error {
	bs.ScanTargetFunc = f
	return nil
}

func (bs *Base) AddObserver(obj BlockScanNotificationObject) error {
	if obj == nil {
		return nil
	}
	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	bs.Observers[obj] = true
	return nil
}

func (bs *Base) RemoveObserver(obj BlockScanNotificationObject) error {
	bs.Mu.Lock()
	defer bs.Mu.Unlock()
	delete(bs.Observers, obj)
	return nil
}

func (bs *Base) SetRescanBlockHeight(height uint64) error { return nil }

func (bs *Base) SetTask(task func()) {
	if bs.taskRunner != nil && bs.taskRunner.Running() {
		bs.taskRunner.Stop()
		bs.taskRunner = nil
	}
	if task != nil {
		bs.taskRunner = &taskRunner{f: task, tick: bs.PeriodOfTask}
	}
}

func (bs *Base) Run() error {
	if bs.isClose {
		return fmt.Errorf("block scanner has been closed")
	}
	if bs.ScanTargetFunc == nil {
		return fmt.Errorf("scan target func is not set")
	}
	if bs.taskRunner == nil {
		return fmt.Errorf("block scanner has not set scan task")
	}
	bs.scanning = true
	bs.taskRunner.Start()
	return nil
}

func (bs *Base) Stop() error {
	if bs.isClose {
		return fmt.Errorf("block scanner has been closed")
	}
	if bs.taskRunner != nil {
		bs.taskRunner.Stop()
	}
	bs.scanning = false
	return nil
}

func (bs *Base) Pause() error {
	if bs.isClose {
		return fmt.Errorf("block scanner has been closed")
	}
	if bs.taskRunner != nil {
		bs.taskRunner.Stop()
	}
	bs.scanning = false
	return nil
}

func (bs *Base) Restart() error {
	if bs.isClose {
		return fmt.Errorf("block scanner has been closed")
	}
	if bs.taskRunner != nil {
		bs.taskRunner.Start()
	}
	bs.scanning = true
	return nil
}

func (bs *Base) IsClose() bool { return bs.isClose }

func (bs *Base) ScanBlock(height uint64) error {
	return fmt.Errorf("ScanBlock not implement")
}

func (bs *Base) GetCurrentBlockHeader() (*types.BlockHeader, error) {
	return nil, fmt.Errorf("GetCurrentBlockHeader not implement")
}

func (bs *Base) GetGlobalMaxBlockHeight() uint64 { return 0 }

func (bs *Base) GetScannedBlockHeight() uint64 { return 0 }

func (bs *Base) ExtractTransactionData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*types.TxExtractData, error) {
	return nil, fmt.Errorf("ExtractTransactionData not implement")
}

func (bs *Base) ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) (map[string][]*types.TxExtractData, map[string]*types.SmartContractReceipt, error) {
	return nil, nil, fmt.Errorf("ExtractTransactionAndReceiptData not implement")
}

func (bs *Base) GetBalanceByAddress(address ...string) ([]*types.Balance, error) {
	return nil, fmt.Errorf("GetBalanceByAddress not implement")
}

func (bs *Base) GetTransactionsByAddress(offset, limit int, coin types.Coin, address ...string) ([]*types.TxExtractData, error) {
	return nil, fmt.Errorf("GetTransactionsByAddress not implement")
}

func (bs *Base) SupportBlockchainDAI() bool { return false }

func (bs *Base) SetBlockchainDAI(dai BlockchainDAI) error {
	bs.BlockchainDAI = dai
	return nil
}

// NewBlockNotify 将新区块推送给观察者。若管道已满则非阻塞丢弃，避免扫块循环被慢观察者阻塞。
func (bs *Base) NewBlockNotify(header *types.BlockHeader) error {
	bs.Mu.RLock()
	defer bs.Mu.RUnlock()
	if bs.isClose {
		return nil
	}
	select {
	case bs.blockProducer <- header:
	default:
		// 管道满时丢弃，避免阻塞扫块协程
	}
	return nil
}

// CloseBlockScanner 先标记关闭再停任务再关管道，避免任务中 NewBlockNotify 向已关闭 channel 写入
func (bs *Base) CloseBlockScanner() error {
	bs.Mu.Lock()
	bs.isClose = true
	bs.Mu.Unlock()
	_ = bs.Stop()
	bs.Mu.Lock()
	close(bs.blockProducer)
	bs.Mu.Unlock()
	return nil
}

func (bs *Base) GetBlockchainSyncStatus() (*types.BlockchainSyncStatus, error) {
	return nil, fmt.Errorf("GetBlockchainSyncStatus not implement")
}
