package signverify

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/godaddy-x/wallet-adapter/types"
)

// Result 签前 message 复现校验结果。
type Result struct {
	OK                bool              `json:"ok"`
	Symbol            string            `json:"symbol"`
	TxType            int64             `json:"txType"`
	Sid               string            `json:"sid"`
	SignScheme        string            `json:"signScheme"`
	ChainID           string            `json:"chainId"`
	MessageExpected   string            `json:"messageExpected"`
	MessageReproduced string            `json:"messageReproduced"`
	SignExt           map[string]string `json:"signExt,omitempty"`
}

type schemeVerifier func(data string, txType int64, signExt map[string]string) (*Result, error)

var schemeVerifiers = make(map[string]schemeVerifier)

// RegisterScheme 注册 signScheme 对应的签前校验实现（如 wallet-adapter-eth 注册 eip155）。
func RegisterScheme(scheme string, fn schemeVerifier) {
	scheme = strings.TrimSpace(scheme)
	if scheme == "" || fn == nil {
		return
	}
	schemeVerifiers[scheme] = fn
}

// VerifyPendingSignData 校验 PendingSignTx.Data：解析 signExt 并按 signScheme 复现 message。
// CLI 须在启动时调用各链 Register（如 eth.RegisterSignVerify）。
func VerifyPendingSignData(data string) (*Result, error) {
	data = strings.TrimSpace(data)
	if data == "" {
		return nil, fmt.Errorf("pending sign data is empty")
	}

	hdr, err := parsePendingHeader(data)
	if err != nil {
		return nil, err
	}
	signExt, err := types.ParseSignExt(hdr.SignExt)
	if err != nil {
		return nil, err
	}
	scheme := types.SignExtScheme(signExt)
	if scheme == "" {
		return nil, fmt.Errorf("signExt.%s is empty", types.SignExtKeySignScheme)
	}
	fn, ok := schemeVerifiers[scheme]
	if !ok {
		return nil, fmt.Errorf("unsupported signScheme: %s", scheme)
	}
	res, err := fn(data, hdr.TxType, signExt)
	if err != nil {
		return nil, err
	}
	if res != nil {
		res.Symbol = hdr.Coin.Symbol
		res.TxType = hdr.TxType
		if res.Sid == "" {
			res.Sid = hdr.Sid
		}
		if res.SignScheme == "" {
			res.SignScheme = scheme
		}
		if res.ChainID == "" {
			res.ChainID = strings.TrimSpace(signExt[types.SignExtKeyChainID])
		}
		if res.SignExt == nil {
			res.SignExt = signExt
		}
	}
	return res, nil
}

type pendingHeader struct {
	TxType   int64  `json:"txType"`
	Sid      string `json:"sid"`
	SignExt  string `json:"signExt"`
	Coin     struct {
		Symbol string `json:"symbol"`
	} `json:"coin"`
}

func parsePendingHeader(data string) (*pendingHeader, error) {
	var hdr pendingHeader
	if err := json.Unmarshal([]byte(data), &hdr); err != nil {
		return nil, fmt.Errorf("pending sign data decode: %w", err)
	}
	return &hdr, nil
}

// NormalizeHex32 去掉 0x 并转小写，便于 message 比对。
func NormalizeHex32(hex string) string {
	hex = strings.TrimSpace(hex)
	hex = strings.TrimPrefix(hex, "0x")
	return strings.ToLower(hex)
}
