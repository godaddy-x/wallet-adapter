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

	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
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

// BuildBatchTransaction 构建批量转账待签单：
// 1) decoder.CreateBatchRawTransaction 构建 rawTx
// 2) 补充 createTime/createNonce/txType 并序列化 RawTransaction
// 3) wrapper.SignPendingTxData 生成 PendingSignTx
func BuildBatchTransaction(d decoder.TransactionDecoder, wrapper wallet.WalletDAI, batch *types.BatchRawRequest) (*types.PendingSignTx, error) {
	if d == nil {
		return nil, errors.New("decoder is nil")
	}
	if batch == nil {
		return nil, errors.New("batch is nil")
	}
	if wrapper == nil {
		return nil, errors.New("wrapper is nil")
	}
	rawTx, err := d.CreateBatchRawTransaction(wrapper, batch)
	if err != nil {
		return nil, err
	}
	if rawTx == nil {
		return nil, errors.New("CreateBatchRawTransaction returned nil rawTx")
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
	originalTxJSON := string(txJSON)

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

	rawTx, err := mergePendingRawTransaction(pendingTx)
	if err != nil {
		return nil, err
	}

	if err := d.VerifyRawTransaction(wrapper, rawTx); err != nil {
		return nil, err
	}

	return d.SubmitRawTransaction(wrapper, rawTx)
}

// BuildSmartContractTransaction 构建合约类原始交易单的 PendingSignTx（与 BuildTransaction 对称）：
// 1) SmartContractDecoder.CreateSmartContractRawTransaction
// 2) 填充 createTime/createNonce/txType（TxType=2 表示合约写链，与 RawTransaction 的 0 单笔 / 1 汇总 区分）
// 3) wrapper.SignPendingTxData 对序列化后的 SmartContractRawTransaction 做业务签
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

// SendSmartContractTransaction 广播合约原始交易单（与 SendTransaction 对称）：
// 校验 dataSign/tradeSign → 反序列化 SmartContractRawTransaction → 合并 SignerList → SubmitSmartContractRawTransaction
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
