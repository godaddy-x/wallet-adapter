package chain

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"

	"github.com/blockchain/wallet-adapter/decoder"
	"github.com/blockchain/wallet-adapter/scanner"
)

// 按 symbol 注册/查询：RegAdapter、GetAdapter、GetTransactionDecoder、GetBlockScanner、GetAddressDecoder、ListSymbols。

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
