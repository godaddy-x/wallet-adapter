package flow

import (
	"errors"
	"fmt"

	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// EvaluateSummaryFeeDeficit re-estimates native fee gap for an existing summary leg at current time.
func EvaluateSummaryFeeDeficit(d decoder.TransactionDecoder, wrapper wallet.WalletDAI, req *types.SummaryFeeDeficitEvalRequest) (*types.SummaryFeeDeficitEvalResult, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if req == nil {
		return nil, errors.New("eval request is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}
	ev, ok := d.(decoder.SummaryFeeDeficitEvaluator)
	if !ok {
		return nil, fmt.Errorf("EvaluateSummaryFeeDeficit not supported for this chain")
	}
	return ev.EvaluateSummaryFeeDeficit(wrapper, req)
}
