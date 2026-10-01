package scanner

import (
	"context"
	"strings"
	"testing"
)

type crossCheckStub struct {
	Base
	enabled bool
	res     CrossCheckResult
	err     error
}

func (c *crossCheckStub) CrossCheckEnabled() bool { return c.enabled }
func (c *crossCheckStub) VerifyBeforePromote(context.Context, CrossCheckRequest) (CrossCheckResult, error) {
	return c.res, c.err
}

func TestRunPromoteCrossCheckMismatchIncludesPeer(t *testing.T) {
	bs := &crossCheckStub{
		enabled: true,
		res: CrossCheckResult{
			OK: false, Reason: "cross_check: blockHash mismatch",
			PeerURL: "https://***@rpc.example/v1",
		},
	}
	err := RunPromoteCrossCheck(context.Background(), bs, CrossCheckRequest{
		TxID: "0x1", AccountID: "acc",
	})
	if err == nil {
		t.Fatal("want error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "peer=https://***@rpc.example/v1") {
		t.Fatalf("want peer in error, got %q", msg)
	}
}
