package flow

import (
	"testing"

	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

type stubSummaryDecoder struct {
	decoder.TransactionDecoderBase
	items []*types.RawTransactionWithError
}

func (s *stubSummaryDecoder) CreateSummaryRawTransactionWithError(_ wallet.WalletDAI, _ *types.SummaryRawTransaction) ([]*types.RawTransactionWithError, error) {
	return s.items, nil
}

type stubSummaryWallet struct {
	wallet.WalletDAIBase
}

func (stubSummaryWallet) SignPendingTxData(txJSON []byte) (*types.PendingSignTx, error) {
	return &types.PendingSignTx{
		Data:      string(txJSON),
		DataSign:  "stub-data-sign",
		TradeSign: "stub-trade-sign",
	}, nil
}

func TestBuildSummaryTransaction_AssignsSidAndFeeDeficits(t *testing.T) {
	rawTx := &types.RawTransaction{
		Coin: types.Coin{Symbol: "TRX"},
		To:   map[string]string{"TSummary": "10"},
	}
	dec := &stubSummaryDecoder{
		items: []*types.RawTransactionWithError{
			{
				RawTx: rawTx,
				FeeDeficit: &types.SummaryFeeDeficit{
					Address:     "TChild0",
					Token:       "USDT",
					TokenAmount: "100",
					Symbol:      "TRX",
					Shortfall:   "10.1",
					Reason:      types.FeeDeficitReasonInsufficientEnergy,
				},
			},
			{
				FeeDeficit: &types.SummaryFeeDeficit{
					Address:   "TChild1",
					Token:     "TRX",
					Symbol:    "TRX",
					Shortfall: "1",
					Reason:    types.FeeDeficitReasonSweepBlocked,
				},
			},
		},
	}
	sum := &types.SummaryRawTransaction{Sid: "abc", Coin: types.Coin{Symbol: "TRX"}}

	got, err := BuildSummaryTransaction(dec, stubSummaryWallet{}, sum)
	if err != nil {
		t.Fatalf("BuildSummaryTransaction: %v", err)
	}
	if len(got.PendingSignTx) != 1 {
		t.Fatalf("pending count=%d want 1", len(got.PendingSignTx))
	}
	if got.PendingSignTx[0].Sid != "abc#s0" {
		t.Fatalf("pending sid=%q", got.PendingSignTx[0].Sid)
	}
	if got.PendingSignTx[0].Message != types.SummaryPendingFeeRechargeRequired {
		t.Fatalf("message=%q", got.PendingSignTx[0].Message)
	}
	if len(got.FeeDeficits) != 2 {
		t.Fatalf("deficit count=%d want 2", len(got.FeeDeficits))
	}
	if got.FeeDeficits[0].Sid != "abc#s0" || got.FeeDeficits[0].Reason != types.FeeDeficitReasonInsufficientEnergy {
		t.Fatalf("deficit[0]=%+v", got.FeeDeficits[0])
	}
	if got.FeeDeficits[1].Sid != "abc#s1" || got.FeeDeficits[1].Reason != types.FeeDeficitReasonSweepBlocked {
		t.Fatalf("deficit[1]=%+v", got.FeeDeficits[1])
	}
}

func TestBuildSummaryTransaction_SweepBlockedOnly(t *testing.T) {
	dec := &stubSummaryDecoder{
		items: []*types.RawTransactionWithError{
			{
				FeeDeficit: &types.SummaryFeeDeficit{
					Address:   "TChild",
					Reason:    types.FeeDeficitReasonSweepBlocked,
					Shortfall: "0.5",
					Symbol:    "TRX",
					Token:     "TRX",
				},
			},
		},
	}
	sum := &types.SummaryRawTransaction{Sid: "batch1", Coin: types.Coin{Symbol: "TRX"}}

	got, err := BuildSummaryTransaction(dec, stubSummaryWallet{}, sum)
	if err != nil {
		t.Fatalf("BuildSummaryTransaction: %v", err)
	}
	if len(got.PendingSignTx) != 0 {
		t.Fatalf("expected no pending legs, got %d", len(got.PendingSignTx))
	}
	if len(got.FeeDeficits) != 1 {
		t.Fatalf("deficit count=%d want 1", len(got.FeeDeficits))
	}
	if got.FeeDeficits[0].Sid != "batch1#s0" {
		t.Fatalf("sid=%q", got.FeeDeficits[0].Sid)
	}
}
