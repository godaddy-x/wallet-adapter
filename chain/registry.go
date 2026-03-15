package chain

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"strings"
	"sync"

	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/scanner"
	"golang.org/x/crypto/ripemd160"
)

// 按 symbol 注册/查询：RegAdapter、GetAdapter、GetTransactionDecoder、GetBlockScanner、GetAddressDecoder、GetSmartContractDecoder、ListSymbols。

var (
	adapters   = make(map[string]ChainAdapter)
	adaptersMu sync.RWMutex
)

// RegAdapter 注册链适配器，symbol 一般为大写（如 BTC、ETH、BTM）
func RegAdapter(symbol string, a ChainAdapter) {
	adaptersMu.Lock()
	defer adaptersMu.Unlock()
	if a == nil {
		panic("adapter is nil")
	}
	adapters[symbol] = a
}

// GetAdapter 按 symbol 获取链适配器
func GetAdapter(symbol string) (ChainAdapter, error) {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	a, ok := adapters[symbol]
	if !ok {
		return nil, fmt.Errorf("adapter not found for symbol: %s", symbol)
	}
	return a, nil
}

// GetTransactionDecoder 按 symbol 获取该链的 TransactionDecoder
func GetTransactionDecoder(symbol string) (decoder.TransactionDecoder, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	d := a.GetTransactionDecoder()
	if d == nil {
		return nil, fmt.Errorf("chain %s does not support transaction build/broadcast", symbol)
	}
	return d, nil
}

// GetBlockScanner 按 symbol 获取该链的 BlockScanner
func GetBlockScanner(symbol string) (scanner.BlockScanner, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	s := a.GetBlockScanner()
	if s == nil {
		return nil, fmt.Errorf("chain %s does not support block scanner", symbol)
	}
	return s, nil
}

// GetAddressDecoder 按 symbol 获取该链的 AddressDecoder
func GetAddressDecoder(symbol string) (decoder.AddressDecoder, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	d := a.GetAddressDecoder()
	if d == nil {
		return nil, fmt.Errorf("chain %s does not support address decoder", symbol)
	}
	return d, nil
}

// GetSmartContractDecoder 按 symbol 获取该链的 SmartContractDecoder（可选，不支持则返回错误）
func GetSmartContractDecoder(symbol string) (decoder.SmartContractDecoder, error) {
	a, err := GetAdapter(symbol)
	if err != nil {
		return nil, err
	}
	d := a.GetSmartContractDecoder()
	if d == nil {
		return nil, fmt.Errorf("chain %s does not support smart contract decoder", symbol)
	}
	return d, nil
}

// ListSymbols 返回已注册的所有 symbol
func ListSymbols() []string {
	adaptersMu.RLock()
	defer adaptersMu.RUnlock()
	symbols := make([]string, 0, len(adapters))
	for s := range adapters {
		symbols = append(symbols, s)
	}
	return symbols
}

// GenContractID 合约ID：symbol_address 的 SHA256 再 Base64 编码
func GenContractID(symbol, address string) string {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("%v_%v", symbol, address)))
	return base64.StdEncoding.EncodeToString(sum[:])
}

// KeyIDVer Base58Check 版本字节，用于 ComputeKeyID
const KeyIDVer = 0x00

// ComputeKeyID 根据 seed 计算 KeyID：SHA256(seed) -> RIPEMD160 -> Base58Check(KeyIDVer)
func ComputeKeyID(seed []byte) string {
	// Step 1: SHA256(seed)
	h := sha256.Sum256(seed)
	// Step 2: RIPEMD160(SHA256(seed))
	rh := ripemd160.New()
	rh.Write(h[:])
	ripemd160Hash := rh.Sum(nil)
	// Step 3: Base58Check 编码（带版本字节）
	return base58CheckEncode(KeyIDVer, ripemd160Hash)
}

// base58CheckEncode 对 payload 做 Base58Check 编码，version 为 1 字节版本号
func base58CheckEncode(version byte, payload []byte) string {
	data := make([]byte, 0, 1+len(payload)+4)
	data = append(data, version)
	data = append(data, payload...)
	checksum := sha256.Sum256(data)
	checksum = sha256.Sum256(checksum[:])
	data = append(data, checksum[:4]...)
	return base58Encode(data)
}

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(data []byte) string {
	// Count leading zeros
	leadingZeros := 0
	for _, b := range data {
		if b != 0 {
			break
		}
		leadingZeros++
	}

	// Convert big-endian bytes to big integer
	num := new(big.Int).SetBytes(data)

	// Encode in base58
	var result []byte
	base := big.NewInt(58)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		result = append(result, base58Alphabet[mod.Int64()])
	}

	// Append leading '1's for each leading zero byte
	for i := 0; i < leadingZeros; i++ {
		result = append(result, '1')
	}

	// Reverse
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
