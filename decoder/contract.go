// Smart contract decoder: token balances, ABI call/create/broadcast, contract metadata; parallel to TransactionDecoder, optional per chain.
package decoder

import (
	"fmt"

	"github.com/godaddy-x/wallet-adapter/types"
	"github.com/godaddy-x/wallet-adapter/wallet"
)

// SmartContractDecoder smart contract decoder interface for chains that support contracts (optional).
// Parallel to TransactionDecoder; ChainAdapter returns it via GetSmartContractDecoder(), which may be nil.
type SmartContractDecoder interface {
	ABIDAI

	// GetTokenBalanceByAddress queries token balance list for addresses.
	GetTokenBalanceByAddress(contract types.SmartContract, address ...string) ([]*types.TokenBalance, error)
	// CallSmartContractABI read-only ABI call; does not produce an on-chain transaction.
	CallSmartContractABI(wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) (*types.SmartContractCallResult, *types.AdapterError)
	// CreateSmartContractRawTransaction creates a smart contract raw transaction (fills Raw, etc.).
	CreateSmartContractRawTransaction(wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) *types.AdapterError
	// SubmitSmartContractRawTransaction broadcasts a smart contract transaction.
	SubmitSmartContractRawTransaction(wrapper wallet.WalletDAI, rawTx *types.SmartContractRawTransaction) (*types.SmartContractReceipt, *types.AdapterError)
	// GetTokenMetadata queries token metadata by contract address.
	GetTokenMetadata(contract string) (*types.SmartContract, error)
}

// ABIDAI ABI data access interface, embedded by SmartContractDecoder.
type ABIDAI interface {
	GetABIInfo(address string) (*types.ABIInfo, error)
	SetABIInfo(address string, abi types.ABIInfo) error
}

// SmartContractDecoderBase default not-implemented base class.
type SmartContractDecoderBase struct{}

func (SmartContractDecoderBase) GetTokenBalanceByAddress(contract types.SmartContract, address ...string) ([]*types.TokenBalance, error) {
	return nil, errNotImplementSC("GetTokenBalanceByAddress")
}

func (SmartContractDecoderBase) CallSmartContractABI(wallet.WalletDAI, *types.SmartContractRawTransaction) (*types.SmartContractCallResult, *types.AdapterError) {
	return nil, types.Errorf(types.ErrSystemException, "CallSmartContractABI not implement")
}

func (SmartContractDecoderBase) CreateSmartContractRawTransaction(wallet.WalletDAI, *types.SmartContractRawTransaction) *types.AdapterError {
	return types.Errorf(types.ErrSystemException, "CreateSmartContractRawTransaction not implement")
}

func (SmartContractDecoderBase) SubmitSmartContractRawTransaction(wallet.WalletDAI, *types.SmartContractRawTransaction) (*types.SmartContractReceipt, *types.AdapterError) {
	return nil, types.Errorf(types.ErrSystemException, "SubmitSmartContractRawTransaction not implement")
}

func (SmartContractDecoderBase) GetABIInfo(address string) (*types.ABIInfo, error) {
	return nil, errNotImplementSC("GetABIInfo")
}

func (SmartContractDecoderBase) SetABIInfo(address string, abi types.ABIInfo) error {
	return errNotImplementSC("SetABIInfo")
}

func (SmartContractDecoderBase) GetTokenMetadata(contract string) (*types.SmartContract, error) {
	return nil, errNotImplementSC("GetTokenMetadata")
}

func errNotImplementSC(method string) error {
	return fmt.Errorf("%s not implement", method)
}
