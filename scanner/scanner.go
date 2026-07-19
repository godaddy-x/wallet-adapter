// Package scanner block scanner interface and base class; scans blocks by height, extracts transactions/receipts, and returns results.
package scanner

import (
	"fmt"
	"sync"
	"time"

	"github.com/godaddy-x/wallet-adapter/types"
)

// BlockScanTargetFunc batch lookup of scan target ownership, used to filter transactions during block scan.
// Caller passes a single batch param (same Symbol + ScanTargetType); callback fills results in place on the object:
//   - hit target: ScanTarget[target] gets non-nil value (address type: accountID string recommended; contract type: *types.Coin recommended)
//   - miss: ScanTarget[target] stays nil
type BlockScanTargetFunc func(target *types.ScanTargetParam) error

// ScanLoopParams parameter struct for RunScanLoop; new params can be added without changing method signature.
type ScanLoopParams struct {
	StartHeight   uint64                           // start scan height; scanning begins at StartHeight+1
	Confirmations uint64                           // confirmation count; only used to compute BlockHeader.Confirmations for business reference
	Interval      time.Duration                    // sleep interval after each scan round
	HandleBlock   func(res *types.BlockScanResult) // callback after each height is scanned (may be nil)
}

// BlockScanner core block scanner interface:
// - scan: scan blocks by height (sync result or error only)
// - verify: second on-chain verification by txid, with strict comparison against business expectations
// - continuous scan: assumes external system maintains cursor; callbacks emit each height's scan result
//
// Notes:
// - Does not include internal persistence (DAI) or business query capabilities (balance/address transactions, etc.); those belong to external systems or separate components.
type BlockScanner interface {
	SetBlockScanTargetFunc(scanTargetFunc BlockScanTargetFunc) error
	// SetTradeOrderLookup wires business outbound snapshot lookup for per-tx fee accounting.
	SetTradeOrderLookup(TradeOrderOutboundQuerier) error

	// Run control: start/stop internal scan task.
	Run() error
	Pause() error

	// ScanBlockWithResult scans a block by height and returns summary result for external cursor advance and retry.
	// Convention: error means "cannot complete scan at this height" (RPC failure, block missing, etc.); result.Success expresses business success/failure.
	ScanBlockWithResult(height uint64) (*types.BlockScanResult, error)
	// ScanBlockOnce scans a height once (for backfill/missed-block repair); no continuous loop or external cursor maintenance.
	ScanBlockOnce(height uint64) (*types.BlockScanResult, error)

	// ResetScanHeight resets continuous scan loop start height (for admin cursor correction/rollback rescan).
	// Convention: only affects running RunScanLoop (if supported); must not restart the process.
	ResetScanHeight(height uint64) error

	GetCurrentBlockHeader() (*types.BlockHeader, error)
	// GetBlockHash lightweight block hash query at height (eth_getBlockByNumber/full=false, etc.) for confirmation-stage newly validity check.
	GetBlockHash(height uint64) (string, error)
	GetGlobalMaxBlockHeight() uint64
	ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) ([]*types.ExtractDataItem, []*types.ContractReceiptItem, error)

	// GetBalanceByAddress queries balance for addresses.
	GetBalanceByAddress(address ...string) ([]*types.Balance, error)

	// VerifyTransactionByTxID pre-credit second on-chain verification by txid returning creditable result set.
	// Convention: error means RPC/system failure preventing verification; business rejection via result.Verified=false + Reason.
	VerifyTransactionByTxID(txid string, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyResult, error)

	// VerifyTransactionMatch pre-credit second verification of on-chain results with strict comparison against external expected.
	// Convention: error for system/RPC faults; business rejection via result.Verified=false + Reason/Mismatches.
	VerifyTransactionMatch(txid string, expected *types.TxVerifyExpected, scanTargetFunc BlockScanTargetFunc, minConfirmations uint64) (*types.TxVerifyMatchResult, error)

	// RunScanLoop continuously scans blocks by height:
	// - starts at params.StartHeight+1, scans upward serially;
	// - scans through latest (chain tip), no longer subtracting Confirmations;
	// - Confirmations only used for BlockHeader.Confirmations field for business reference;
	// - each height scanned once, no duplicate scans to save resources;
	// - after each height, calls params.HandleBlock (if non-nil) to callback ScanBlockWithResult to external system;
	// - sleeps Interval after each round then loops.
	// This method only produces candidate results; credit/confirm/retry policy decided externally via callbacks and Verify APIs.
	RunScanLoop(params ScanLoopParams) error

	// ScanBlockPrioritize priority scan for specified height list.
	// Notes:
	// - enqueues priority heights; RunScanLoop processes them between main-line scans;
	// - priority scan results reuse RunScanLoop params.HandleBlock (no callback if RunScanLoop not running);
	// - priority scan does not affect RunScanLoop main cursor advancement;
	// - priority heights processed ascending, deduplicated;
	// - all heights must satisfy params.Confirmations (height <= latest - params.Confirmations) or error is returned;
	// - caller may adjust heights per error and retry.
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

// Base block scanner base class: provides ScanTargetFunc injection, task run control, and default not-implemented methods.
// Chain implementations should embed this struct and override BlockScanner methods as needed.
type Base struct {
	Mu             sync.RWMutex
	ScanTargetFunc BlockScanTargetFunc
	// TradeOrderLookup optional; one GetTradeOrderOutbound call per tx during fee_extract.
	TradeOrderLookup TradeOrderOutboundQuerier

	PeriodOfTask time.Duration
	taskRunner   *taskRunner
}

// NewBlockScannerBase creates scanner base
func NewBlockScannerBase() *Base {
	return &Base{PeriodOfTask: defaultPeriodOfTask}
}

func (bs *Base) SetBlockScanTargetFunc(f BlockScanTargetFunc) error {
	bs.ScanTargetFunc = f
	return nil
}

// SetTradeOrderLookup wires business trade order lookup (e.g. open_scanner ScanWrapper).
func (bs *Base) SetTradeOrderLookup(q TradeOrderOutboundQuerier) error {
	bs.TradeOrderLookup = q
	return nil
}

// SetTask sets internal periodic task (not exposed on interface; injected by chain scanners at construction).
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

func (bs *Base) GetBlockHash(height uint64) (string, error) {
	return "", fmt.Errorf("GetBlockHash not implement")
}

func (bs *Base) GetGlobalMaxBlockHeight() uint64 { return 0 }

func (bs *Base) ExtractTransactionAndReceiptData(txid string, scanTargetFunc BlockScanTargetFunc) ([]*types.ExtractDataItem, []*types.ContractReceiptItem, error) {
	return nil, nil, fmt.Errorf("ExtractTransactionAndReceiptData not implement")
}

// GetBalanceByAddress returns not-implemented error; implemented by chain-specific scanners.
// Chain implementations should call QueryBalancesConcurrent helper for concurrent queries.
func (bs *Base) GetBalanceByAddress(address ...string) ([]*types.Balance, error) {
	return nil, fmt.Errorf("GetBalanceByAddress not implement")
}

// BalanceQueryFunc callback to query single address balance, invoked by QueryBalancesConcurrent.
type BalanceQueryFunc func(address string) (confirmed, unconfirmed, total string, err error)

// QueryBalancesConcurrent concurrently queries balances for multiple addresses.
// Chain scanners should call this helper when implementing GetBalanceByAddress.
//
// Parameters:
//   - symbol: chain identifier
//   - addresses: address list
//   - query: single-address query callback returning confirmed, unconfirmed, and total balance
//   - concurrency: concurrency limit, default 20
//
// Returns Balance list in the same order as input addresses.
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
		result     = make([]*types.Balance, len(addresses))
		sem        = make(chan struct{}, concurrency)
		resultChan = make(chan *addrResult, len(addresses))
		errChan    = make(chan error, 1)
		wg         sync.WaitGroup
	)

	// result collector goroutine
	go func() {
		for r := range resultChan {
			result[r.index] = r.balance
		}
	}()

	// concurrent queries
	for i, addr := range addresses {
		wg.Add(1)
		go func(idx int, address string) {
			defer wg.Done()

			sem <- struct{}{}        // acquire semaphore
			defer func() { <-sem }() // release semaphore

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

	// check for errors
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

func (bs *Base) RunScanLoop(params ScanLoopParams) error {
	return fmt.Errorf("RunScanLoop not implement")
}

// ScanBlockPrioritize default returns not-implemented error; chain scanners should override.
// Implementation reference: while RunScanLoop runs, enqueue priority heights; RunScanLoop handles them between main-line scans.
func (bs *Base) ScanBlockPrioritize(heights []uint64) error {
	return fmt.Errorf("ScanBlockPrioritize not implement")
}
