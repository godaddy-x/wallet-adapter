package scanner

import (
	"context"
	"fmt"
	"strings"
)

// CrossCheckRequest first-scan anchor for Phase C (mapped from ow_trade_newly).
//
// outputIndex is chain-specific (see open_gateway docs/TRADE_OUTPUT_INDEX.md). Examples:
//   ETH: -1 main coin value leg; >=0 + contract = ERC20 log index; -2 fee.
//   BTC: >=0 vout receive; -3 per-payer net (disambiguate with NativeFrom / From on newly); -2 fee; not -1 for typical deposit.
// Leg strict-match rules live in each chain's CrossCheckValidator (e.g. wallet-adapter-eth), not in this struct.
type CrossCheckRequest struct {
	Symbol      string
	MainSymbol  string // chain family hint: BTC, ETH, TRX, …
	TxID        string
	BlockHeight uint64
	BlockHash   string
	AccountID   string
	BlockTime   int64
	Source      string
	TxAction    string // send | receive | internal | fee | activation | …

	ContractAddress string
	OutputIndex     int64 // semantics vary by MainSymbol; always pass through from newly

	NativeFrom     string
	NativeTo       string
	NativeAmount   string
	NativeDecimals int32

	TokenFrom     string
	TokenTo       string
	TokenAmount   string
	TokenDecimals int32
}

// CrossCheckResult cross-source verification outcome.
type CrossCheckResult struct {
	OK        bool
	Reason    string // mismatch: prefer prefix cross_check:
	PeerIndex int
	PeerURL   string // redacted identifier for audit
}

// CrossCheckValidator Phase C: re-verify promote conclusion using verifyAPIs peer pool (not main scan RPC).
type CrossCheckValidator interface {
	CrossCheckEnabled() bool
	VerifyBeforePromote(ctx context.Context, req CrossCheckRequest) (CrossCheckResult, error)
}

// RunPromoteCrossCheck invokes Phase C when bs implements CrossCheckValidator and peers are configured.
func RunPromoteCrossCheck(ctx context.Context, bs BlockScanner, req CrossCheckRequest) error {
	if bs == nil {
		return nil
	}
	v, ok := bs.(CrossCheckValidator)
	if !ok || !v.CrossCheckEnabled() {
		return nil
	}
	res, err := v.VerifyBeforePromote(ctx, req)
	if err != nil {
		return fmt.Errorf("promote cross_check: %w", err)
	}
	if res.OK {
		return nil
	}
	return crossCheckPromoteError(req, res)
}

func crossCheckPromoteError(req CrossCheckRequest, res CrossCheckResult) error {
	reason := strings.TrimSpace(res.Reason)
	if reason == "" {
		reason = "cross_check: mismatch"
	}
	if peer := strings.TrimSpace(res.PeerURL); peer != "" {
		return fmt.Errorf("%s txID=%s accountID=%s peer=%s", reason, req.TxID, req.AccountID, peer)
	}
	return fmt.Errorf("%s txID=%s accountID=%s", reason, req.TxID, req.AccountID)
}
