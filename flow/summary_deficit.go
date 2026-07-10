package flow

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/godaddy-x/wallet-adapter/amount"
	"github.com/godaddy-x/wallet-adapter/types"
)

// TokenDisplaySymbol returns the token label for summary feeDeficits.
func TokenDisplaySymbol(contract types.SmartContract) string {
	if s := strings.TrimSpace(contract.Token); s != "" {
		return s
	}
	return strings.TrimSpace(contract.Symbol)
}

// FormatNativeShortfall formats native coin shortfall for SummaryFeeDeficit.Shortfall.
func FormatNativeShortfall(shortfall *big.Int, nativeDecimal int32) string {
	if shortfall == nil || shortfall.Sign() <= 0 {
		return "0"
	}
	s, err := amount.BigIntToDecimal(shortfall, int64(nativeDecimal))
	if err != nil {
		return "0"
	}
	return s
}

// NewSummaryFeeDeficit builds a deficit record from chain-unit shortfall.
func NewSummaryFeeDeficit(addr, token, tokenAmount, nativeSymbol, reason string, shortfall *big.Int, nativeDecimal int32) *types.SummaryFeeDeficit {
	if shortfall == nil || shortfall.Sign() <= 0 {
		return nil
	}
	return &types.SummaryFeeDeficit{
		Address:     addr,
		Token:       token,
		TokenAmount: tokenAmount,
		Symbol:      nativeSymbol,
		Shortfall:   FormatNativeShortfall(shortfall, nativeDecimal),
		Reason:      reason,
	}
}

// NewNativeGasFeeDeficit builds insufficient_native_for_gas deficit for token summary legs.
func NewNativeGasFeeDeficit(addr, token, tokenAmount, nativeSymbol string, nativeBal, fee *big.Int, nativeDecimal int32) *types.SummaryFeeDeficit {
	if nativeBal == nil {
		nativeBal = big.NewInt(0)
	}
	if fee == nil || fee.Sign() <= 0 || nativeBal.Cmp(fee) >= 0 {
		return nil
	}
	shortfall := new(big.Int).Sub(fee, nativeBal)
	return NewSummaryFeeDeficit(addr, token, tokenAmount, nativeSymbol, types.FeeDeficitReasonInsufficientNativeForGas, shortfall, nativeDecimal)
}

// NewSweepBlockedFeeDeficit reports native summary blocked by retained balance and fees.
func NewSweepBlockedFeeDeficit(addr, nativeSymbol string, addrBal, sumAmount *big.Int, nativeDecimal int32) *types.SummaryFeeDeficit {
	if addrBal == nil || sumAmount == nil || sumAmount.Sign() > 0 {
		return nil
	}
	gap := new(big.Int).Neg(sumAmount)
	if gap.Sign() <= 0 {
		return nil
	}
	tokenAmount, _ := amount.BigIntToDecimal(addrBal, int64(nativeDecimal))
	return &types.SummaryFeeDeficit{
		Address:     addr,
		Token:       nativeSymbol,
		TokenAmount: tokenAmount,
		Symbol:      nativeSymbol,
		Shortfall:   FormatNativeShortfall(gap, nativeDecimal),
		Reason:      types.FeeDeficitReasonSweepBlocked,
	}
}

// BuildSummaryFeeDeficitEvalResult maps a live deficit estimate to API response fields.
func BuildSummaryFeeDeficitEvalResult(req *types.SummaryFeeDeficitEvalRequest, deficit *types.SummaryFeeDeficit, needRecreate bool) *types.SummaryFeeDeficitEvalResult {
	res := &types.SummaryFeeDeficitEvalResult{
		SummarySid:          req.SummarySid,
		PreviousShortfall:   req.PreviousShortfall,
		NeedRecreateSummary: needRecreate,
		CanRetryBroadcast:   true,
	}
	if deficit == nil {
		res.CanRetryBroadcast = !needRecreate
		return res
	}
	res.Address = deficit.Address
	res.Token = deficit.Token
	res.TokenAmount = deficit.TokenAmount
	res.Symbol = deficit.Symbol
	res.Shortfall = deficit.Shortfall
	res.Reason = deficit.Reason
	if deficit.Reason == types.FeeDeficitReasonSweepBlocked {
		res.CanRetryBroadcast = false
	} else {
		res.CanRetryBroadcast = !needRecreate
	}
	return res
}

// DecodeSummaryRawTx parses summary leg RawTransaction from pendingSignTx.data.
func DecodeSummaryRawTx(pending *types.PendingSignTx) (*types.RawTransaction, error) {
	if pending == nil || strings.TrimSpace(pending.Data) == "" {
		return nil, fmt.Errorf("pendingSignTx.data is empty")
	}
	rawTx := &types.RawTransaction{}
	if err := json.Unmarshal([]byte(pending.Data), rawTx); err != nil {
		return nil, fmt.Errorf("decode summary raw tx failed: %w", err)
	}
	return rawTx, nil
}

// SummarySourceAddress returns the child address funding a summary leg.
func SummarySourceAddress(rawTx *types.RawTransaction) string {
	if rawTx == nil || len(rawTx.TxFrom) == 0 {
		return ""
	}
	part := strings.SplitN(rawTx.TxFrom[0], ":", 2)[0]
	return strings.TrimSpace(part)
}

// SummaryAmountFromRaw returns the sweep amount encoded on the summary leg.
func SummaryAmountFromRaw(rawTx *types.RawTransaction) string {
	if rawTx == nil {
		return ""
	}
	if rawTx.TxAmount != "" {
		return rawTx.TxAmount
	}
	if len(rawTx.TxFrom) > 0 {
		parts := strings.SplitN(rawTx.TxFrom[0], ":", 2)
		if len(parts) == 2 {
			return parts[1]
		}
	}
	for _, v := range rawTx.To {
		return v
	}
	return ""
}

// SummaryTargetAddress returns the summary destination address.
func SummaryTargetAddress(rawTx *types.RawTransaction) string {
	if rawTx == nil {
		return ""
	}
	for k := range rawTx.To {
		return k
	}
	if len(rawTx.TxTo) > 0 {
		return strings.SplitN(rawTx.TxTo[0], ":", 2)[0]
	}
	return ""
}
