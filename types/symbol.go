package types

// BalanceModelType balance model (by address or by account).
type BalanceModelType uint32

const (
	BalanceModelTypeAddress BalanceModelType = 0
	BalanceModelTypeAccount BalanceModelType = 1
)

// SymbolInfo chain/coin info interface (Symbol, Decimal, CurveType, FullName, BalanceModelType).
type SymbolInfo interface {
	Symbol() string
	Decimal() int32
	CurveType() uint32
	FullName() string
	BalanceModelType() BalanceModelType
}

// SymbolInfoBase default empty SymbolInfo implementation for ChainAdapterBase embedding.
type SymbolInfoBase struct{}

func (SymbolInfoBase) Symbol() string                    { return "" }
func (SymbolInfoBase) Decimal() int32                    { return 0 }
func (SymbolInfoBase) CurveType() uint32                  { return 0 }
func (SymbolInfoBase) FullName() string                  { return "" }
func (SymbolInfoBase) BalanceModelType() BalanceModelType { return BalanceModelTypeAddress }
