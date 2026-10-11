package scanner

import (
	"sync"
	"time"
)

// LoopPause coordinates Pause/Run for RunScanLoop and background pollers (e.g. SOL tipPoller).
// Must be heap-allocated via NewLoopPause so sync.Cond stays bound to this mutex.
type LoopPause struct {
	mu      sync.Mutex
	cond    *sync.Cond
	paused  bool
	pauseCh chan struct{}
}

// NewLoopPause returns an initially running pause gate.
func NewLoopPause() *LoopPause {
	lp := &LoopPause{pauseCh: make(chan struct{})}
	lp.cond = sync.NewCond(&lp.mu)
	return lp
}

func (lp *LoopPause) Run() {
	if lp == nil {
		return
	}
	lp.mu.Lock()
	if lp.paused {
		lp.paused = false
		lp.pauseCh = make(chan struct{})
		lp.cond.Broadcast()
	}
	lp.mu.Unlock()
}

func (lp *LoopPause) Pause() {
	if lp == nil {
		return
	}
	lp.mu.Lock()
	if !lp.paused {
		lp.paused = true
		if lp.pauseCh != nil {
			close(lp.pauseCh)
		}
	}
	lp.mu.Unlock()
}

func (lp *LoopPause) ScanPaused() bool {
	if lp == nil {
		return false
	}
	lp.mu.Lock()
	p := lp.paused
	lp.mu.Unlock()
	return p
}

func (lp *LoopPause) BlockIfPaused() {
	if lp == nil {
		return
	}
	lp.mu.Lock()
	for lp.paused {
		lp.cond.Wait()
	}
	lp.mu.Unlock()
}

func (lp *LoopPause) PauseCh() <-chan struct{} {
	if lp == nil {
		ch := make(chan struct{})
		return ch
	}
	lp.mu.Lock()
	ch := lp.pauseCh
	if ch == nil {
		ch = make(chan struct{})
		lp.pauseCh = ch
	}
	lp.mu.Unlock()
	return ch
}

// WaitInterruptible sleeps for d or until Pause closes PauseCh.
func (lp *LoopPause) WaitInterruptible(d time.Duration) {
	if lp == nil {
		time.Sleep(d)
		return
	}
	select {
	case <-time.After(d):
	case <-lp.PauseCh():
	}
}
