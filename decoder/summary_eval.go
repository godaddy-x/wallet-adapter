package decoder

import (
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// SummaryFeeDeficitEvaluator optional extension for chains that support live summary fee re-evaluation.
type SummaryFeeDeficitEvaluator interface {
	EvaluateSummaryFeeDeficit(wrapper wallet.WalletDAI, req *types.SummaryFeeDeficitEvalRequest) (*types.SummaryFeeDeficitEvalResult, error)
}
