package flow

import (
	"fmt"

	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/mailru/easyjson"
)

func mergePendingRawTransaction(pendingTx *types.PendingSignTx) (*types.RawTransaction, error) {
	if pendingTx == nil || pendingTx.Data == "" {
		return nil, fmt.Errorf("pendingTx.data is nil")
	}
	rawTx := &types.RawTransaction{}
	if err := easyjson.Unmarshal([]byte(pendingTx.Data), rawTx); err != nil {
		return nil, fmt.Errorf("rawTx json error: %w", err)
	}
	for accountID, keySignatures := range rawTx.Signatures {
		if keySignatures != nil {
			for k, keySignature := range keySignatures {
				keySignature.Signature = pendingTx.SignerList[fmt.Sprintf("%s-%d", accountID, k)]
			}
		}
		rawTx.Signatures[accountID] = keySignatures
	}
	return rawTx, nil
}

func mergePendingSmartContractRawTransaction(pendingTx *types.PendingSignTx) (*types.SmartContractRawTransaction, error) {
	if pendingTx == nil || pendingTx.Data == "" {
		return nil, fmt.Errorf("pendingTx.data is nil")
	}
	rawTx := &types.SmartContractRawTransaction{}
	if err := easyjson.Unmarshal([]byte(pendingTx.Data), rawTx); err != nil {
		return nil, fmt.Errorf("smart contract rawTx json error: %w", err)
	}
	for accountID, keySignatures := range rawTx.Signatures {
		if keySignatures != nil {
			for k, keySignature := range keySignatures {
				sigKey := fmt.Sprintf("%s-%d", accountID, k)
				sign, ok := pendingTx.SignerList[sigKey]
				if !ok || sign == "" {
					return nil, fmt.Errorf("pendingTx.signerList missing key %q", sigKey)
				}
				keySignature.Signature = sign
			}
		}
		rawTx.Signatures[accountID] = keySignatures
	}
	return rawTx, nil
}
