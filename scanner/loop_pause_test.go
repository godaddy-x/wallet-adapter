package scanner

import (
	"sync"
	"testing"
	"time"
)

func TestLoopPausePauseBlocksAndRunResumes(t *testing.T) {
	lp := NewLoopPause()
	if lp.ScanPaused() {
		t.Fatal("expected initially running")
	}

	lp.Pause()
	if !lp.ScanPaused() {
		t.Fatal("expected paused")
	}

	done := make(chan struct{})
	go func() {
		lp.BlockIfPaused()
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("BlockIfPaused returned while paused")
	case <-time.After(50 * time.Millisecond):
	}

	lp.Run()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("BlockIfPaused did not resume after Run")
	}
	if lp.ScanPaused() {
		t.Fatal("expected running after Run")
	}
}

func TestLoopPausePauseChInterruptsWait(t *testing.T) {
	lp := NewLoopPause()
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		lp.WaitInterruptible(2 * time.Second)
	}()
	time.Sleep(20 * time.Millisecond)
	lp.Pause()
	wg.Wait()
	if time.Since(start) > 500*time.Millisecond {
		t.Fatalf("WaitInterruptible not interrupted by Pause: %v", time.Since(start))
	}
}

func TestLoopPauseIdempotent(t *testing.T) {
	lp := NewLoopPause()
	lp.Pause()
	lp.Pause()
	if !lp.ScanPaused() {
		t.Fatal("expected paused")
	}
	lp.Run()
	lp.Run()
	if lp.ScanPaused() {
		t.Fatal("expected running")
	}
}
