// Package flow transaction build and broadcast flow; signing is provided by external MPC; this package builds PendingSignTx and submits broadcasts.
package flow

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
	"github.com/mailru/easyjson"
)

// GetRandomSecure generates a byte slice of the given length using a cryptographically secure RNG (recommended).
func GetRandomSecure(l int) ([]byte, error) {
	randomIV := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, randomIV); err != nil {
		return nil, err
	}
	return randomIV, nil
}

// BuildTransaction builds a normal transaction:
// 1) decoder builds rawTx (rawHex/fees/sigParts, etc.)
// 2) flow fills createTime/createNonce/txType and serializes to JSON
// 3) wrapper.SignPendingTxData signs the rawTx JSON on the business side, returning PendingSignTx (with dataSign/tradeSign)
// Important: PendingSignTx.Data must use a copy of originalTxJSON so wrapper-internal mutations of txJSON do not affect final submit data
func BuildTransaction(d decoder.TransactionDecoder, wrapper wallet.WalletDAI, rawTx *types.RawTransaction) (*types.PendingSignTx, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if rawTx == nil {
		return nil, errors.New("rawTx is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}
	if err := d.CreateRawTransaction(wrapper, rawTx); err != nil {
		return nil, err
	}
	nonce, err := GetRandomSecure(32)
	if err != nil {
		return nil, err
	}
	rawTx.CreateTime = time.Now().UnixMilli()
	rawTx.CreateNonce = hex.EncodeToString(nonce)
	rawTx.TxType = 0

	txJSON, err := easyjson.Marshal(rawTx)
	if err != nil {
		return nil, err
	}
	originalTxJSON := string(txJSON) // required: copy so callback mutations of txJSON do not affect the transaction payload

	signData, err := wrapper.SignPendingTxData(txJSON)
	if err != nil {
		return nil, err
	}
	signData.Data = originalTxJSON
	signData.Sid = rawTx.Sid
	return signData, nil
}

// BuildSummaryTransaction builds a list of summary transactions (each may carry Error):
// decoder returns RawTransactionWithError list, then each entry gets nonce/time/type filled -> wrapper signs -> PendingSignTx is produced.
func BuildSummaryTransaction(d decoder.TransactionDecoder, wrapper wallet.WalletDAI, sumRawTx *types.SummaryRawTransaction) ([]*types.PendingSignTx, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if sumRawTx == nil {
		return nil, errors.New("sumRawTx is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}

	rawTxArray, err := d.CreateSummaryRawTransactionWithError(wrapper, sumRawTx)
	if err != nil {
		return nil, fmt.Errorf("CreateSummaryRawTransactionWithError error: %w", err)
	}
	if len(rawTxArray) == 0 {
		return nil, errors.New("CreateSummaryRawTransactionWithError create is nil")
	}

	now := time.Now().UnixMilli()
	txData := make([]*types.PendingSignTx, 0, len(rawTxArray))
	for k, v := range rawTxArray {
		nonce, err := GetRandomSecure(32)
		if err != nil {
			return nil, err
		}
		rawTx := v.RawTx
		if rawTx == nil {
			continue
		}
		rawTx.Sid = fmt.Sprintf("%s#%d", sumRawTx.Sid, k)
		rawTx.CreateTime = now
		rawTx.CreateNonce = hex.EncodeToString(nonce)
		rawTx.TxType = 1

		txJSON, err := easyjson.Marshal(rawTx)
		if err != nil {
			return nil, err
		}
		originalTxJSON := string(txJSON) // required: copy so callback mutations of txJSON do not affect the transaction payload

		signData, err := wrapper.SignPendingTxData(txJSON)
		if err != nil {
			return nil, err
		}
		signData.Data = originalTxJSON
		signData.Sid = rawTx.Sid
		if v.Error != nil {
			signData.Code = strconv.FormatUint(v.Error.Code, 10)
			signData.Message = v.Error.Error()
		}
		txData = append(txData, signData)
	}
	return txData, nil
}

// SendTransaction broadcasts a transaction:
// 1) verify dataSign/tradeSign match Data to prevent tampering
// 2) deserialize Data to rawTx, merge SignerList -> rawTx.Signatures
// 3) VerifyRawTransaction validates signatures
// 4) SubmitRawTransaction broadcasts
func SendTransaction(d decoder.TransactionDecoder, wrapper wallet.WalletDAI, pendingTx *types.PendingSignTx) (*types.Transaction, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}
	if pendingTx == nil {
		return nil, errors.New("pendingTx is nil")
	}
	if pendingTx.Data == "" {
		return nil, errors.New("pendingTx.data is nil")
	}
	if pendingTx.DataSign == "" {
		return nil, errors.New("pendingTx.dataSign is nil")
	}
	if pendingTx.TradeSign == "" {
		return nil, errors.New("pendingTx.tradeSign is nil")
	}
	if len(pendingTx.SignerList) == 0 {
		return nil, errors.New("pendingTx.signerList is nil")
	}

	checkData, err := wrapper.SignPendingTxData([]byte(pendingTx.Data))
	if err != nil {
		return nil, err
	}
	if checkData.DataSign != pendingTx.DataSign || checkData.TradeSign != pendingTx.TradeSign {
		return nil, errors.New("pendingTx.dataSign or pendingTx.tradeSign invalid")
	}

	rawTx, err := mergePendingRawTransaction(pendingTx)
	if err != nil {
		return nil, err
	}

	if err := d.VerifyRawTransaction(wrapper, rawTx); err != nil {
		return nil, err
	}

	return d.SubmitRawTransaction(wrapper, rawTx)
}

// BuildSmartContractTransaction builds PendingSignTx for a contract raw transaction (symmetric to BuildTransaction):
// 1) SmartContractDecoder.CreateSmartContractRawTransaction
// 2) fill createTime/createNonce/txType (TxType=2 means contract write; distinct from RawTransaction 0 single / 1 summary)
// 3) wrapper.SignPendingTxData signs the serialized SmartContractRawTransaction on the business side
func BuildSmartContractTransaction(d decoder.SmartContractDecoder, wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) (*types.PendingSignTx, error) {
	if d == nil {
		return nil, errors.New("smart contract decoder is nil")
	}
	if rawTx == nil {
		return nil, errors.New("rawTx is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}
	if ae := d.CreateSmartContractRawTransaction(wrapper, rawTx); ae != nil {
		return nil, ae
	}
	nonce, err := GetRandomSecure(32)
	if err != nil {
		return nil, err
	}
	rawTx.CreateTime = time.Now().UnixMilli()
	rawTx.CreateNonce = hex.EncodeToString(nonce)
	rawTx.TxType = 2

	txJSON, err := easyjson.Marshal(rawTx)
	if err != nil {
		return nil, err
	}
	originalTxJSON := string(txJSON)

	signData, err := wrapper.SignPendingTxData(txJSON)
	if err != nil {
		return nil, err
	}
	signData.Data = originalTxJSON
	signData.Sid = rawTx.Sid
	return signData, nil
}

// SendSmartContractTransaction broadcasts a contract raw transaction (symmetric to SendTransaction):
// verify dataSign/tradeSign -> deserialize SmartContractRawTransaction -> merge SignerList -> SubmitSmartContractRawTransaction
func SendSmartContractTransaction(d decoder.SmartContractDecoder, wrapper wallet.WalletDAI, pendingTx *types.PendingSignTx) (*types.SmartContractReceipt, error) {
	if d == nil {
		return nil, errors.New("smart contract decoder is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}
	if pendingTx == nil {
		return nil, errors.New("pendingTx is nil")
	}
	if pendingTx.Data == "" {
		return nil, errors.New("pendingTx.data is nil")
	}
	if pendingTx.DataSign == "" {
		return nil, errors.New("pendingTx.dataSign is nil")
	}
	if pendingTx.TradeSign == "" {
		return nil, errors.New("pendingTx.tradeSign is nil")
	}
	if len(pendingTx.SignerList) == 0 {
		return nil, errors.New("pendingTx.signerList is nil")
	}

	checkData, err := wrapper.SignPendingTxData([]byte(pendingTx.Data))
	if err != nil {
		return nil, err
	}
	if checkData.DataSign != pendingTx.DataSign || checkData.TradeSign != pendingTx.TradeSign {
		return nil, errors.New("pendingTx.dataSign or pendingTx.tradeSign invalid")
	}

	rawTx, err := mergePendingSmartContractRawTransaction(pendingTx)
	if err != nil {
		return nil, err
	}

	receipt, ae := d.SubmitSmartContractRawTransaction(wrapper, rawTx)
	if ae != nil {
		return receipt, ae
	}
	return receipt, nil
}
