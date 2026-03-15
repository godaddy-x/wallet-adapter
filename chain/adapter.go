// Package chain 链适配器接口与注册表，提供 ChainAdapter 及按 symbol 的 RegAdapter/GetAdapter/GetTransactionDecoder/GetBlockScanner/GetAddressDecoder。
package chain

import (
	"github.com/blockchain/wallet-adapter/decoder"
	"github.com/blockchain/wallet-adapter/scanner"
	"github.com/blockchain/wallet-adapter/types"
)

// ChainAdapter 链适配器接口：多主链统一入口，聚合 SymbolInfo、AssetsConfig、TransactionDecoder、BlockScanner、AddressDecoder；可选 SmartContractDecoder。
type ChainAdapter interface {
	types.SymbolInfo
	AssetsConfig

	GetTransactionDecoder() decoder.TransactionDecoder
	GetBlockScanner() scanner.BlockScanner
	GetAddressDecoder() decoder.AddressDecoder
	// GetSmartContractDecoder 智能合约解析器，可选；不支持合约的链返回 nil
	GetSmartContractDecoder() decoder.SmartContractDecoder
}

// ChainAdapterBase 基类，GetTransactionDecoder / GetBlockScanner / GetAddressDecoder / GetSmartContractDecoder 返回 nil
type ChainAdapterBase struct {
	types.SymbolInfoBase
	AssetsConfigBase
}

func (ChainAdapterBase) GetTransactionDecoder() decoder.TransactionDecoder {
	return nil
}

func (ChainAdapterBase) GetBlockScanner() scanner.BlockScanner {
	return nil
}

func (ChainAdapterBase) GetAddressDecoder() decoder.AddressDecoder {
	return nil
}

func (ChainAdapterBase) GetSmartContractDecoder() decoder.SmartContractDecoder {
	return nil
}
