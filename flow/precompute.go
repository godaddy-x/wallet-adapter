package flow

import (
	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

type rawTxIDPrecomputer interface {
	PrecomputeSubmitTxID(wallet.WalletDAI, *types.RawTransaction) (string, error)
}

type smartContractTxIDPrecomputer interface {
	PrecomputeSmartContractSubmitTxID(wallet.WalletDAI, *types.SmartContractRawTransaction) (string, error)
}

func PrecomputeTransactionTxID(d decoder.TransactionDecoder, wrapper wallet.WalletDAI, pendingTx *types.PendingSignTx) (string, error) {
	if d == nil || pendingTx == nil {
		return "", nil
	}
	rawTx, err := mergePendingRawTransaction(pendingTx)
	if err != nil {
		return "", err
	}
	pc, ok := d.(rawTxIDPrecomputer)
	if !ok {
		return "", nil
	}
	return pc.PrecomputeSubmitTxID(wrapper, rawTx)
}

func PrecomputeSmartContractTransactionTxID(d decoder.SmartContractDecoder, wrapper wallet.WalletDAI, pendingTx *types.PendingSignTx) (string, error) {
	if d == nil || pendingTx == nil {
		return "", nil
	}
	rawTx, err := mergePendingSmartContractRawTransaction(pendingTx)
	if err != nil {
		return "", err
	}
	pc, ok := d.(smartContractTxIDPrecomputer)
	if !ok {
		return "", nil
	}
	return pc.PrecomputeSmartContractSubmitTxID(wrapper, rawTx)
}
