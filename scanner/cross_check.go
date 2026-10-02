package scanner

import (
	"context"
	"fmt"
	"strings"
)

// CrossCheckRequest first-scan anchor for Phase C (mapped from newly).
// All chains: tx existence + block anchor.
// ETH native (NativeStrict): from/to/value on eth_getTransactionByHash.
// ETH token (TokenStrict): chain adapters may use Token* + eth_getTransactionReceipt Transfer log (optional until enabled per chain).
type CrossCheckRequest struct {
	Symbol      string
	TxID        string
	BlockHeight uint64 // EVM/TRX height; SOL maps slot into this field at mapper if needed
	BlockHash   string // ETH/BTC required when chain validates hash; SOL C2 may use slot-only rules
	AccountID   string // audit / error messages only
	BlockTime   int64  // SOL optional auxiliary
	Source      string // auto_promote | manual_approve
	// NativeStrict: EVM simple native leg (e.g. ETH OutputIndex=-1, non-contract). Compares peer tx from/to/value to newly.
	NativeStrict   bool
	NativeFrom     string
	NativeTo       string
	NativeAmount   string // decimal amount string (same semantics as newly.Amount)
	NativeDecimals int32  // e.g. 18 for ETH
	// TokenStrict: contract token leg (e.g. ETH ERC20 OutputIndex>=0). Populated from newly for future receipt/log cross-check.
	TokenStrict          bool
	TokenContractAddress string // newly.ContractAddress
	TokenLogIndex        int64  // newly.OutputIndex (ERC20 log index)
	TokenFrom            string
	TokenTo              string
	TokenAmount   string // decimal string (newly.Amount / ToAddressV)
	TokenDecimals int32  // newly.Decimals
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
// Unconfigured or unsupported chains return nil (no-op).
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
