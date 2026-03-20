// Package scanner 区块扫描器接口与基类，负责按高度扫描区块、提取交易/回执并返回结果。
package scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/godaddy-x/wallet-adapter/types"
)

// BlockScanTargetFunc 根据扫描目标参数（地址/别名等）查询所属源与是否存在，供扫块时过滤交易。
type BlockScanTargetFunc func(target types.ScanTargetParam) types.ScanTargetResult

// TokenMetadataFunc 根据链标识与合约地址查询代币/合约元数据（SmartContract），供扫块器在提取交易/回执时补充合约信息。
type TokenMetadataFunc func(symbol, contractAddr string) *types.SmartContract

// BlockScanner 区块扫描器核心接口：
// - 扫块：按高度扫描区块（同步返回结果或仅返回 error）
// - 复核：按 txid 二次链上复核，并可与业务期望严格比对
// - 持续扫描：以“外部系统维护游标”为前提，回调输出每个高度的扫描结果
//
// 说明：
// - 不包含内部持久化（DAI）与业务查询类能力（余额/地址交易等），这些应由外部系统或独立组件负责。
type BlockScanner interface {
	SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error

	SetTokenMetadataFunc(tokenMetadataFunc TokenMetadataFunc) error

	// 运行控制：启动/停止内部扫描任务。
	Run() error
	Pause() error

	// ScanBlockWithResult 按高度扫描区块并返回摘要结果，供外部系统推进游标与重试。
	// 约定：error 用于表达“无法完成该高度扫描”（如 RPC 失败/块不存在等）；result.Success 表达业务语义上的成功与否。
	ScanBlockWithResult(height uint64) (*types.BlockScanResult, error)
	// ScanBlockOnce 指定高度扫描一次（用于补扫/漏扫修复），不进入持续循环、不维护外部游标。
	ScanBlockOnce(height uint64) (*types.BlockScanResult, error)

	// ResetScanHeight 将“持续扫块循环”的起始高度重置到指定值（用于后台指令修正游标/回滚重扫）。
	// 约定：该方法只影响正在运行的 RunScanLoop（若实现方支持），不应引起进程重启。
	ResetScanHeight(height uint64) error

	GetCurrentBlockHeader() (*types.BlockHeader, error)
	GetGlobalMaxBlockHeight() uint64
	ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) ([]*types.ExtractDataItem, []*types.ContractReceiptItem, error)

	// GetBalanceByAddress 查询指定地址的余额。
	GetBalanceByAddress(address ...string) ([]*types.Balance, error)

	// VerifyTransactionByTxID 入账前按 txid 二次复核链上结果并返回可入账结果集。
	// 约定：error 用于表达“RPC/系统错误导致无法完成复核”；业务层面的不通过以 result.Verified=false + Reason 表达。
	VerifyTransactionByTxID(txid string, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyResult, error)

	// VerifyTransactionMatch 入账前对链上结果集做二次复核，并与外部期望对象 expected 严格比对。
	// 约定：error 表达系统/RPC 异常；业务不通过以 result.Verified=false + Reason/Mismatches 表达。
	VerifyTransactionMatch(txid string, expected *types.TxVerifyExpected, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyMatchResult, error)

	// RunScanLoop 按高度持续扫描区块：
	// - 从 startHeight+1 开始，串行向上扫描；
	// - 每轮根据当前 latest 与 confirmations 计算 safeTo=latest-confirmations，仅扫描至 safeTo；
	// - 为减少重组影响，可通过 windowSize 控制回填窗口（当 startHeight 早于 safeTo-windowSize 时从窗口起点开始重扫）；
	// - 每扫完一个高度调用 handleBlock（若非空）将 ScanBlockWithResult 的结果同步回调给外部系统；
	// - 每轮结束后 sleep interval 再次循环。
	// 该方法只负责生产候选结果，入账/确认/重试策略由外部系统基于回调结果与 Verify 接口自行决定。
	RunScanLoop(startHeight, confirmations, windowSize uint64, interval time.Duration, handleBlock func(res *types.BlockScanResult)) error

	// ScanBlockPrioritize 插队扫描指定高度列表。
	// 说明：
	// - 将插队高度加入优先队列，RunScanLoop 会在主线扫描间隙优先处理这些高度；
	// - 插队高度的扫描结果复用 RunScanLoop 的 handleBlock（如果 RunScanLoop 未运行则无回调）；
	// - 插队扫描不影响 RunScanLoop 的主线 cursor 推进逻辑；
	// - 插队高度按升序处理，且去重；
	// - 严格要求所有传入的高度都满足 confirmations 要求（height <= latest - confirmations），否则直接返回错误；
	// - 调用方可根据错误信息调整高度后重试。
	ScanBlockPrioritize(heights []uint64) error
}

const defaultPeriodOfTask = 5 * time.Second

type taskRunner struct {
	f    func()
	tick time.Duration
	stop chan struct{}
}

func (t *taskRunner) Start() {
	if t == nil || t.f == nil || t.tick <= 0 {
		return
	}
	if t.stop != nil {
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
	if t == nil || t.stop == nil {
		return
	}
	close(t.stop)
	t.stop = nil
}

func (t *taskRunner) Running() bool { return t != nil && t.stop != nil }

// Base 区块扫描器基类：提供 ScanTargetFunc/TokenMetadataFunc 注入、任务运行控制与默认未实现方法。
// 各链实现建议嵌入该结构体，并按需重写 BlockScanner 接口中的方法。
type Base struct {
	Mu                sync.RWMutex
	ScanTargetFunc    BlockScanTargetFunc
	TokenMetadataFunc TokenMetadataFunc

	PeriodOfTask time.Duration
	taskRunner   *taskRunner
}

// NewBlockScannerBase 创建扫描器基类
func NewBlockScannerBase() *Base {
	return &Base{PeriodOfTask: defaultPeriodOfTask}
}

func (bs *Base) SetBlockScanTargetFunc(f BlockScanTargetFunc) error {
	bs.ScanTargetFunc = f
	return nil
}

func (bs *Base) SetTokenMetadataFunc(f TokenMetadataFunc) error {
	bs.TokenMetadataFunc = f
	return nil
}

// SetTask 设置内部周期任务（不在接口中暴露，供各链扫描器在构造时注入）。
func (bs *Base) SetTask(task func()) {
	if bs.taskRunner != nil && bs.taskRunner.Running() {
		bs.taskRunner.Stop()
	}
	if task == nil {
		bs.taskRunner = nil
		return
	}
	tick := bs.PeriodOfTask
	if tick <= 0 {
		tick = defaultPeriodOfTask
	}
	bs.taskRunner = &taskRunner{f: task, tick: tick}
}

func (bs *Base) Run() error {
	if bs.taskRunner == nil {
		return fmt.Errorf("block scanner has not set scan task")
	}
	bs.taskRunner.Start()
	return nil
}

func (bs *Base) Stop() error {
	if bs.taskRunner != nil {
		bs.taskRunner.Stop()
	}
	return nil
}

func (bs *Base) Pause() error { return bs.Stop() }

func (bs *Base) Restart() error { return bs.Run() }

func (bs *Base) ScanBlockWithResult(height uint64) (*types.BlockScanResult, error) {
	return nil, fmt.Errorf("ScanBlockWithResult not implement")
}

func (bs *Base) ScanBlockOnce(height uint64) (*types.BlockScanResult, error) {
	return nil, fmt.Errorf("ScanBlockOnce not implement")
}

func (bs *Base) ResetScanHeight(height uint64) error {
	return fmt.Errorf("ResetScanHeight not implement")
}

func (bs *Base) GetCurrentBlockHeader() (*types.BlockHeader, error) {
	return nil, fmt.Errorf("GetCurrentBlockHeader not implement")
}

func (bs *Base) GetGlobalMaxBlockHeight() uint64 { return 0 }

func (bs *Base) ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) ([]*types.ExtractDataItem, []*types.ContractReceiptItem, error) {
	return nil, nil, fmt.Errorf("ExtractTransactionAndReceiptData not implement")
}

// GetBalanceByAddress 返回未实现错误，由具体链的扫描器实现。
// 各链实现应调用 QueryBalancesConcurrent 辅助函数进行并发查询。
func (bs *Base) GetBalanceByAddress(address ...string) ([]*types.Balance, error) {
	return nil, fmt.Errorf("GetBalanceByAddress not implement")
}

// BalanceQueryFunc 查询单个地址余额的回调函数，由 QueryBalancesConcurrent 调用。
type BalanceQueryFunc func(address string) (confirmed, unconfirmed, total string, err error)

// QueryBalancesConcurrent 并发查询多个地址余额。
// 各链扫描器在实现 GetBalanceByAddress 时调用此辅助方法。
//
// 参数：
//   - symbol: 链标识
//   - addresses: 地址列表
//   - query: 单地址查询回调函数，返回已确认余额、未确认余额、总余额
//   - concurrency: 并发限制，默认20
//
// 返回按传入地址顺序排列的 Balance 列表。
func (bs *Base) QueryBalancesConcurrent(symbol string, addresses []string, query BalanceQueryFunc, concurrency int) ([]*types.Balance, error) {
	if len(addresses) == 0 {
		return make([]*types.Balance, 0), nil
	}
	if concurrency <= 0 {
		concurrency = 20
	}

	type addrResult struct {
		index   int
		balance *types.Balance
	}

	var (
		result      = make([]*types.Balance, len(addresses))
		sem         = make(chan struct{}, concurrency)
		resultChan  = make(chan *addrResult, len(addresses))
		errChan     = make(chan error, 1)
		wg          sync.WaitGroup
	)

	// 结果收集协程
	go func() {
		for r := range resultChan {
			result[r.index] = r.balance
		}
	}()

	// 并发查询
	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address string) {
			defer wg.Done()

			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			confirmed, unconfirmed, total, err := query(address)
			if err != nil {
				select {
				case errChan <- fmt.Errorf("query balance for address %s failed: %w", address, err):
				default:
				}
				return
			}

			resultChan <- &addrResult{
				index: idx,
				balance: &types.Balance{
					Symbol:           symbol,
					Address:          address,
					ConfirmBalance:   confirmed,
					UnconfirmBalance: unconfirmed,
					Balance:          total,
				},
			}
		}(i, addr)
	}

	wg.Wait()
	close(resultChan)

	// 检查是否有错误
	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	return result, nil
}

func (bs *Base) VerifyTransactionByTxID(txid string, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyResult, error) {
	return nil, fmt.Errorf("VerifyTransactionByTxID not implement")
}

func (bs *Base) VerifyTransactionMatch(txid string, expected *types.TxVerifyExpected, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyMatchResult, error) {
	return nil, fmt.Errorf("VerifyTransactionMatch not implement")
}

func (bs *Base) RunScanLoop(startHeight, confirmations, windowSize uint64, interval time.Duration, handleBlock func(res *types.BlockScanResult)) error {
	return fmt.Errorf("RunScanLoop not implement")
}

// ScanBlockPrioritize 默认返回未实现错误，各链扫描器应重写此方法。
// 实现参考：在 RunScanLoop 运行时，将插队高度加入优先队列，由 RunScanLoop 在主线间隙处理。
func (bs *Base) ScanBlockPrioritize(heights []uint64) error {
	return fmt.Errorf("ScanBlockPrioritize not implement")
}
