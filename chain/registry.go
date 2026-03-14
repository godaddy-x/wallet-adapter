package chain

import (
	"fmt"
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
