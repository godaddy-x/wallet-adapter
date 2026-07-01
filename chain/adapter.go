// Package chain chain adapter interface and registry, providing ChainAdapter and RegAdapter/GetAdapter/GetTransactionDecoder/GetBlockScanner/GetAddressDecoder/GetSmartContractDecoder/ListSymbols by symbol.
package chain

import (
	"github.com/godaddy-x/wallet-adapter/decoder"
	"github.com/godaddy-x/wallet-adapter/scanner"
	"github.com/godaddy-x/wallet-adapter/types"
)

// ChainAdapter chain adapter interface: unified multi-chain entry aggregating SymbolInfo, AssetsConfig, TransactionDecoder, BlockScanner, AddressDecoder; SmartContractDecoder is optional.
type ChainAdapter interface {
	types.SymbolInfo
	AssetsConfig

	GetTransactionDecoder() decoder.TransactionDecoder
	GetBlockScanner() scanner.BlockScanner
	GetAddressDecoder() decoder.AddressDecoder
	// GetSmartContractDecoder smart contract decoder, optional; chains without contract support return nil
	GetSmartContractDecoder() decoder.SmartContractDecoder
}

// ChainAdapterBase base class; GetTransactionDecoder / GetBlockScanner / GetAddressDecoder / GetSmartContractDecoder return nil
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
