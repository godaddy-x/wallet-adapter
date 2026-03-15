// Package flow 交易构建与广播流程；签名由外部 MPC 提供，本包负责构建待签交易单与提交广播。
package flow

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/blockchain/wallet-adapter/decoder"
	"github.com/blockchain/wallet-adapter/types"
	"github.com/blockchain/wallet-adapter/wallet"
	"github.com/mailru/easyjson"
)

// GetRandomSecure 使用加密安全的随机数生成器生成指定字节数组（推荐）
func GetRandomSecure(l int) ([]byte, error) {
	randomIV := make([]byte, l)
	if _, err := io.ReadFull(rand.Reader, randomIV); err != nil {
		return nil, err
	}
	return randomIV, nil
}

// BuildTransaction 构建普通交易单：
// 1) decoder 负责构建 rawTx（rawHex/fees/sigParts 等）
// 2) flow 负责补充 createTime/createNonce/txType 并序列化为 JSON
// 3) wrapper.SignPendingTxData 对 rawTx JSON 做业务侧签名，返回 PendingSignTx（含 dataSign/tradeSign）
// 重要：PendingSignTx.Data 必须使用 originalTxJSON 副本，避免 wrapper 内部修改 txJSON 影响最终提交数据
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
	originalTxJSON := string(txJSON) // 重要必须：获得副本可无视回调函数修改 txJSON 交易单

	signData, err := wrapper.SignPendingTxData(txJSON)
	if err != nil {
		return nil, err
	}
	signData.Data = originalTxJSON
	signData.Sid = rawTx.Sid
	return signData, nil
}

// BuildSummaryTransaction 构建汇总交易单列表（每笔可能带 Error）：
// 由 decoder 返回 RawTransactionWithError 列表，然后逐笔补齐 nonce/time/type -> wrapper 签名 -> 生成 PendingSignTx。
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
		originalTxJSON := string(txJSON) // 重要必须：获得副本可无视回调函数修改 txJSON 交易单

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

// SendTransaction 广播交易单：
// 1) 校验 dataSign/tradeSign 与 Data 一致，防止 Data 被篡改
// 2) Data 反序列化 rawTx，合并 SignerList -> rawTx.Signatures
// 3) VerifyRawTransaction 校验签名
// 4) SubmitRawTransaction 广播
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

	if err := d.VerifyRawTransaction(wrapper, rawTx); err != nil {
		return nil, err
	}

	return d.SubmitRawTransaction(wrapper, rawTx)
}
