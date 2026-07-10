package flow

import (
	"math/big"
	"testing"

	"github.com/godaddy-x/wallet-adapter/types"
)

func TestNewSweepBlockedFeeDeficit(t *testing.T) {
	deficit := NewSweepBlockedFeeDeficit("TAddr", "TRX", big.NewInt(5_000_000), big.NewInt(-1_000_000), 6)
	if deficit == nil {
		t.Fatal("expected deficit")
	}
	if deficit.Reason != types.FeeDeficitReasonSweepBlocked {
		t.Fatalf("reason=%s", deficit.Reason)
	}
	if deficit.Shortfall != "1" {
		t.Fatalf("shortfall=%q", deficit.Shortfall)
	}
}

func TestNewNativeGasFeeDeficit(t *testing.T) {
	deficit := NewNativeGasFeeDeficit("0xabc", "USDT", "100", "ETH", big.NewInt(1), big.NewInt(10), 18)
	if deficit == nil {
		t.Fatal("expected deficit")
	}
	if deficit.Reason != types.FeeDeficitReasonInsufficientNativeForGas {
		t.Fatalf("reason=%s", deficit.Reason)
	}
}

func TestBuildSummaryFeeDeficitEvalResult_SweepBlocked(t *testing.T) {
	req := &types.SummaryFeeDeficitEvalRequest{SummarySid: "abc#s0", PreviousShortfall: "1"}
	res := BuildSummaryFeeDeficitEvalResult(req, &types.SummaryFeeDeficit{Reason: types.FeeDeficitReasonSweepBlocked}, false)
	if res.CanRetryBroadcast {
		t.Fatal("sweep_blocked should not allow retry broadcast")
	}
}

func TestSummaryLegParse(t *testing.T) {
	raw := &types.RawTransaction{
		TxAmount: "50",
		TxFrom:   []string{"TChild:50"},
		To:       map[string]string{"TSummary": "50"},
	}
	if SummarySourceAddress(raw) != "TChild" {
		t.Fatalf("source=%q", SummarySourceAddress(raw))
	}
	if SummaryAmountFromRaw(raw) != "50" {
		t.Fatalf("amount=%q", SummaryAmountFromRaw(raw))
	}
	if SummaryTargetAddress(raw) != "TSummary" {
		t.Fatalf("target=%q", SummaryTargetAddress(raw))
	}
}
