package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// SignExt 标准 key：链 adapter 建单时写入 signExt JSON，CLI 签前校验使用。
const (
	SignExtKeyChainID           = "chainId"
	SignExtKeySignScheme        = "signScheme"
	SignExtKeyUnsignedEncoding  = "unsignedEncoding"
	SignExtKeyCurveType         = "curveType"
	SignExtKeyHashAlgorithm     = "hashAlgorithm"
)

const (
	SignExtSchemeEIP155     = "eip155"
	SignExtEncodingRLP      = "rlp"
	SignExtHashKeccak256    = "keccak256"
)

// BuildSignExtJSON 将键值对序列化为 signExt JSON 字符串。
func BuildSignExtJSON(fields map[string]string) (string, error) {
	if len(fields) == 0 {
		return "", fmt.Errorf("signExt fields is empty")
	}
	b, err := json.Marshal(fields)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParseSignExt 解析 signExt JSON 为 map。
func ParseSignExt(signExt string) (map[string]string, error) {
	signExt = strings.TrimSpace(signExt)
	if signExt == "" {
		return nil, fmt.Errorf("signExt is empty")
	}
	out := make(map[string]string)
	if err := json.Unmarshal([]byte(signExt), &out); err != nil {
		return nil, fmt.Errorf("signExt json decode: %w", err)
	}
	return out, nil
}

// SignExtChainID 从 signExt map 读取 chainId（十进制字符串）。
func SignExtChainID(ext map[string]string) (uint64, error) {
	if ext == nil {
		return 0, fmt.Errorf("signExt map is nil")
	}
	raw := strings.TrimSpace(ext[SignExtKeyChainID])
	if raw == "" {
		return 0, fmt.Errorf("signExt.%s is empty", SignExtKeyChainID)
	}
	chainID, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("signExt.%s invalid: %w", SignExtKeyChainID, err)
	}
	return chainID, nil
}

// SignExtScheme 返回 signScheme，缺省为空字符串。
func SignExtScheme(ext map[string]string) string {
	if ext == nil {
		return ""
	}
	return strings.TrimSpace(ext[SignExtKeySignScheme])
}
